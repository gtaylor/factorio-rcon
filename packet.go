package rcon

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
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
	maxPacketSize         = 4096 + packetPaddingSize + packetHeaderFieldSize*3
)

var (
	ErrMergeNoPackets   = errors.New("rcon: no packets to merge")
	ErrMergeInvalidID   = errors.New("rcon: mismatched packet ID in merge")
	ErrMergeInvalidType = errors.New("rcon: mismatched packet type in merge")
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

func MergePackets(packets []*Packet) (*Packet, error) {
	var size, id, typ int32
	var body bytes.Buffer

	// if the slice is empty, we can't do anything
	if len(packets) < 1 {
		return nil, ErrMergeNoPackets
	}

	// if we only have one packet, there's no reason to merge anything
	if len(packets) == 1 {
		return packets[0], nil
	}

	// merge packets, taking care to throw errors if the merge doesn't make sense
	for _, packet := range packets {
		if id != 0 && packet.ID != id {
			return nil, ErrMergeInvalidID
		}
		if typ != 0 && packet.Type != typ {
			return nil, ErrMergeInvalidType
		}

		size += packet.Size - packetHeaderSize - packetPaddingSize
		id = packet.ID
		typ = packet.Type
		body.WriteString(packet.Body)
	}

	// return a new merged packet
	return &Packet{size, id, typ, body.String()}, nil
}
