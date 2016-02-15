package rcon

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
)

const (
	terminationSequence        = "\x00"
	failedAuthResponseID int32 = -1
)

var (
	ErrInvalidWrite        = errors.New("rcon: failed to write to remote connection")
	ErrInvalidRead         = errors.New("rcon: failed to read from remote connection")
	ErrInvalidID           = errors.New("rcon: invalid response ID from remote connection")
	ErrInvalidAuthResponse = errors.New("rcon: invalid response type during auth")
	ErrAuthFailed          = errors.New("rcon: authentication failed")
)

type RemoteConsole struct {
	// TODO: add some more useful stuff here?
	Address string
	conn    net.Conn
}

// func (r *RemoteConsole) LocalAddr() net.Addr {
// 	return r.conn.LocalAddr()
// }

// func (r *RemoteConsole) RemoteAddr() net.Addr {
// 	return r.conn.RemoteAddr()
// }

func Dial(address string) (*RemoteConsole, error) {
	// dial tcp
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	// create remote console
	rc := &RemoteConsole{Address: address, conn: conn}
	return rc, nil
}

func (r *RemoteConsole) Close() error {
	return r.conn.Close()
}

func (r *RemoteConsole) Execute(command string) (response *Packet, err error) {
	// Send command to execute
	cmd := NewPacket(ExecCommand, command)
	if err = r.WritePacket(cmd); err != nil {
		return
	}

	// Send "sentinel" packet, to detect if the response has been split.
	// We'll send an empty ResponseValue packet to the server, which will get
	// a new id. The server will respond back with another empty ResponseValue.
	// Because it the server always answers in the order that it received
	// packets, we can use this empty packet to determine when we're done with
	// fetching the response. This approach is documented here:
	// https://developer.valvesoftware.com/wiki/Source_RCON_Protocol#Multiple-packet_Responses
	sentinel := NewPacket(ResponseValue, "")
	if err = r.WritePacket(sentinel); err != nil {
		return
	}

	// Get responses until we hit the sentinel value
	responses := []*Packet{}
	for {
		response, err = r.ReadPacket()
		if err != nil {
			return
		}
		if response.Header.ID == cmd.Header.ID {
			responses = append(responses, response)
		} else {
			break
		}
	}

	// Merge responses into one packet
	response, err = MergePackets(responses)
	if err != nil {
		return
	}
	return
}

func (r *RemoteConsole) Authenticate(password string) (err error) {
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
	if response.Header.ID != packet.Header.ID {
		return ErrInvalidID
	}

	// The server will potentiall send a blank ResponseValue packet before giving
	// back the correct AuthResponse. This can safely be discarded, as documented here:
	// https://developer.valvesoftware.com/wiki/Source_RCON_Protocol#SERVERDATA_AUTH_RESPONSE
	if response.Header.Type == ResponseValue {
		response, err = r.ReadPacket()
		if err != nil {
			return err
		}
	}

	// By now we should for sure have an AuthResponse. If we don't, there's something weird
	// going on server-side
	if response.Header.Type != AuthResponse {
		panic("WTF!?")
	}

	// Check that we did not receive an ID indicating that authentication failed.
	if response.Header.ID == failedAuthResponseID {
		return ErrAuthFailed
	}
	return
}

func (r *RemoteConsole) WritePacket(packet *Packet) (err error) {
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

func (r *RemoteConsole) ReadPacket() (response *Packet, err error) {
	// Read header fields into Header struct
	var header Header
	if err = binary.Read(r.conn, binary.LittleEndian, &header.Size); err != nil {
		return
	}
	if err = binary.Read(r.conn, binary.LittleEndian, &header.ID); err != nil {
		return
	}
	if err = binary.Read(r.conn, binary.LittleEndian, &header.Type); err != nil {
		return
	}

	// Read rest of packet
	var n int
	bytesRead := 0
	bytesTotal := int(header.Size - packetHeaderSize)
	buf := make([]byte, bytesTotal)

	for bytesRead < bytesTotal {
		n, err = r.conn.Read(buf[bytesRead:])
		if err != nil {
			return
		}
		bytesRead += n
	}

	// Trim null bytes off body
	body := strings.TrimRight(string(buf), terminationSequence)

	// Construct final response packet
	response = &Packet{
		Header: header,
		Body:   body,
	}
	return
}
