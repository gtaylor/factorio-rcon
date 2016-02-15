package rcon

import (
	"bytes"
	"testing"
)

func TestNewPacket(t *testing.T) {
	p := NewPacket(Auth, "password")

	if p.Size != 18 {
		t.Error("Expected packet size 18, got ", p.Size)
	}
	if p.Type != Auth {
		t.Error("Expected packet type Auth(3), got ", p.Type)
	}
	if p.Body != "password" {
		t.Error("Expected packet body \"password\", got ", p.Body)
	}
}

func TestPayload(t *testing.T) {
	p := NewPacket(Auth, "password")
	payload, _ := p.Payload()

	size := payload[0:4]
	typ := payload[8:12]
	body := payload[12 : len(payload)-2]
	padding := payload[len(payload)-2:]

	if !bytes.Equal(size, []byte{18, 0, 0, 0}) {
		t.Error("Expected payload [0:4] to be bytes [18 0 0 0], got ", payload[0:4])
	}
	if !bytes.Equal(typ, []byte{3, 0, 0, 0}) {
		t.Error("Expected payload [8:12] to be bytes [3 0 0 0], got ", typ)
	}
	if !bytes.Equal(body, []byte("password")) {
		t.Error("Expected payload body to be bytes \"password\", got ", body)
	}
	if !bytes.Equal(padding, []byte("\x00\x00")) {
		t.Error("Expected two bytes of null padding at end of payload, got ", padding)
	}
}
