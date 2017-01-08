package rcon

import (
	"encoding/binary"
	"errors"
	"net"
	"strings"
)

const (
	terminationSequence        = "\x00"
	failedAuthResponseID int32 = -1
)

var (
	ErrInvalidWrite       = errors.New("rcon: failed to write to remote connection")
	ErrInvalidID          = errors.New("rcon: invalid response ID from remote connection")
	ErrInvalidPacketOrder = errors.New("rcon: packets from server received out of order")
	ErrAuthFailed         = errors.New("rcon: authentication failed")
)

type RCON struct {
	// TODO: add some more useful stuff here?
	Address string
	conn    net.Conn
}

func Dial(address string) (*RCON, error) {
	// dial tcp
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	// create remote console
	rc := &RCON{Address: address, conn: conn}
	return rc, nil
}

func (r *RCON) Close() error {
	return r.conn.Close()
}

func (r *RCON) Execute(command string) (response *Packet, err error) {
	// Send command to execute
	cmd := NewPacket(ExecCommand, command)
	if err = r.WritePacket(cmd); err != nil {
		return
	}

	response, err = r.ReadPacket()
	if err != nil {
		return
	}

	// Handle sentinel package
	if response.ID == cmd.ID {
		// append responses with same id
		return
	} else {
		// something has gotten out of order
		return nil, ErrInvalidPacketOrder
	}
}

func (r *RCON) Authenticate(password string) (err error) {
	// Send auth package
	packet := NewPacket(Auth, password)
	if err = r.WritePacket(packet); err != nil {
		return
	}

	// Get response
	var response *Packet
	response, err = r.ReadPacket()
	if err != nil {
		return
	}
	// Check that response returned correct ID
	if response.ID != packet.ID {
		return ErrInvalidID
	}

	// The server will potentially send a blank ResponseValue packet before giving
	// back the correct AuthResponse. This can safely be discarded, as documented here:
	// https://developer.valvesoftware.com/wiki/Source_RCON_Protocol#SERVERDATA_AUTH_RESPONSE
	if response.Type == ResponseValue {
		response, err = r.ReadPacket()
		if err != nil {
			return err
		}
	}

	// By now we should for sure have an AuthResponse. If we don't, there's something weird
	// going on server-side
	if response.Type != AuthResponse {
		panic("WTF!?")
	}

	// Check that we did not receive an ID indicating that authentication failed.
	if response.ID == failedAuthResponseID {
		return ErrAuthFailed
	}
	return
}

func (r *RCON) WritePacket(packet *Packet) (err error) {
	// generate payload
	var payload []byte
	payload, err = packet.Payload()
	if err != nil {
		return
	}

	// write payload to tcp socket
	var n int
	n, err = r.conn.Write(payload)
	if err != nil {
		return
	}
	if n != len(payload) {
		return ErrInvalidWrite
	}
	return
}

func (r *RCON) ReadPacket() (response *Packet, err error) {
	// Read header fields into Packet struct
	var packet Packet
	if err = binary.Read(r.conn, binary.LittleEndian, &packet.Size); err != nil {
		return
	}
	if err = binary.Read(r.conn, binary.LittleEndian, &packet.ID); err != nil {
		return
	}
	if err = binary.Read(r.conn, binary.LittleEndian, &packet.Type); err != nil {
		return
	}

	// Read rest of packet
	var n int
	bytesRead := 0
	bytesTotal := int(packet.Size - packetHeaderSize)
	buf := make([]byte, bytesTotal)

	for bytesRead < bytesTotal {
		n, err = r.conn.Read(buf[bytesRead:])
		if err != nil {
			return
		}
		bytesRead += n
	}

	// Trim null bytes off body
	packet.Body = strings.TrimRight(string(buf), terminationSequence)

	// Construct final response packet
	return &packet, nil
}
