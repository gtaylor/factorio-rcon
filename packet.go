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
	packetOverhead        = packetPaddingSize + packetHeaderFieldSize*3
	maxPacketSize         = 4096 + packetOverhead
)

type Header struct {
	Size int32
	ID   int32
	Type int32
}

type Packet struct {
	Header Header
	Body   string
}

func NewPacket(typ int32, body string) *Packet {
	var size, id int32

	// calculate size
	size = int32(len(body) + packetHeaderSize + packetPaddingSize)

	// assign a random request id
	binary.Read(rand.Reader, binary.LittleEndian, &id)

	// return packet
	return &Packet{Header{size, id, typ}, body}
}

func (p *Packet) Payload() (payload []byte, err error) {
	buffer := bytes.NewBuffer(make([]byte, 0, p.Header.Size+int32(4)))

	// write size
	binary.Write(buffer, binary.LittleEndian, &p.Header.Size)

	// write request id
	binary.Write(buffer, binary.LittleEndian, &p.Header.ID)

	// write type
	binary.Write(buffer, binary.LittleEndian, &p.Header.Type)

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
		return nil, errors.New("must have at least one packet")
	}

	// if we only have one packet, there's no reason to merge anything
	if len(packets) == 1 {
		return packets[0], nil
	}

	// merge packets, taking care to throw errors if the merge doesn't make sense
	for _, packet := range packets {
		if id != 0 && packet.Header.ID != id {
			return nil, errors.New("mismatched header ids")
		}
		if typ != 0 && packet.Header.Type != typ {
			return nil, errors.New("mismatched packet types")
		}

		size += packet.Header.Size - packetHeaderSize - packetPaddingSize
		id = packet.Header.ID
		typ = packet.Header.Type
		body.WriteString(packet.Body)
	}

	// return a new merged packet
	return &Packet{Header{size, id, typ}, body.String()}, nil
}
