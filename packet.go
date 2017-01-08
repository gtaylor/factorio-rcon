package rcon

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
)

const (
	Auth          int32 = 3
	AuthResponse  int32 = 2
	ExecCommand   int32 = 2
	ResponseValue int32 = 0
)

const (
	packetPaddingSize     = 2
	packetHeaderFieldSize = 4
	packetHeaderSize      = packetHeaderFieldSize * 2
)

type Packet struct {
	Size int32
	ID   int32
	Type int32
	Body string
}

func NewPacket(typ int32, body string) *Packet {
	var size, id int32

	// calculate size
	size = int32(len(body) + packetHeaderSize + packetPaddingSize)

	// assign a random request id
	binary.Read(rand.Reader, binary.LittleEndian, &id)

	// return packet
	return &Packet{size, id, typ, body}
}

func (p *Packet) Payload() (payload []byte, err error) {
	buffer := bytes.NewBuffer(make([]byte, 0, p.Size+packetHeaderFieldSize))

	// write header fields
	binary.Write(buffer, binary.LittleEndian, p.Size)
	binary.Write(buffer, binary.LittleEndian, p.ID)
	binary.Write(buffer, binary.LittleEndian, p.Type)

	// write null-terminated string
	buffer.WriteString(p.Body)
	binary.Write(buffer, binary.LittleEndian, byte(0))

	// write padding
	binary.Write(buffer, binary.LittleEndian, byte(0))

	return buffer.Bytes(), err
}
