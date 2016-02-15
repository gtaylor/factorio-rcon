package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
)

const (
	terminationSequence        = "\x00"
	failedAuthResponseID int32 = -1
)

// const (
// 	readBufferSize      = 4110
// )

var (
	ErrInvalidWrite        = errors.New("rcon: failed to write to remote connection")
	ErrInvalidRead         = errors.New("rcon: failed to read from remote connection")
	ErrInvalidID           = errors.New("rcon: invalid response ID from remote connection")
	ErrInvalidAuthResponse = errors.New("rcon: invalid response type during auth")
	ErrAuthFailed          = errors.New("rcon: authentication failed")
)

type RemoteConsole struct {
	Address string
	conn    net.Conn
}

// func (r *RemoteConsole) LocalAddr() net.Addr {
// 	return r.conn.LocalAddr()
// }

// func (r *RemoteConsole) RemoteAddr() net.Addr {
// 	return r.conn.RemoteAddr()
// }

func Dial(address, password string) (*RemoteConsole, error) {
	// dial tcp
	println("Dialing tcp")
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

func (r *RemoteConsole) Execute(typ int32, str string) (*Packet, error) {
	println("creating packet")
	packet := NewPacket(typ, str)

	println("writing packet")
	err := r.WritePacket(packet)
	if err != nil {
		return nil, err
	}

	println("reading response")
	return nil, nil
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
	println("payload:")
	fmt.Println(payload)
	fmt.Println(hex.EncodeToString(payload))

	// write payload to tcp socket
	var n int
	n, err = r.conn.Write(payload)
	if err != nil {
		return
	}
	if n != len(payload) {
		err = ErrInvalidWrite
		return
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

	// Read body
	var n int
	body := make([]byte, header.Size-packetHeaderSize)
	n, err = r.conn.Read(body)
	if err != nil {
		return
	}
	if n != len(body) {
		err = ErrInvalidRead
		return
	}

	// Construct final response packet
	response = &Packet{
		Header: header,
		Body:   strings.TrimRight(string(body), terminationSequence),
	}
	return
}

func main() {
	r, err := Dial("csgo.steelseries.io:27015", "suckseries")
	if err != nil {
		panic(err)
	}
	defer r.Close()

	err = r.Authenticate("suckseries")
	if err != nil {
		panic(err)
	}
}
