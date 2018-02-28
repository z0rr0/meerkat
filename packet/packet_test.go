package packet

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
)

func compare(a, b *Packet) error {
	if a.ServiceID != b.ServiceID {
		return errors.New("mismatch ServiceID")
	}
	if !bytes.Equal(a.ClientID, b.ClientID) {
		return errors.New("mismatch ClientID")
	}
	if !bytes.Equal(a.Payload, b.Payload) {
		return errors.New("mismatch Payload")
	}
	return nil
}

func TestEncodeDecode(t *testing.T) {
	payload := make([]byte, 256)
	clientID := make([]byte, hashSize)

	if _, err := rand.Read(payload); err != nil {
		t.Fatal(err)
	}
	if _, err := rand.Read(clientID); err != nil {
		t.Fatal(err)
	}

	a := &Packet{ServiceID: 123, ClientID: clientID, Payload: payload}
	data := Encode(a)
	if len(data) == 0 {
		t.Error("empty encoded bytes slice")
	}

	b := Decode(data)
	if err := compare(a, b); err != nil {
		t.Errorf("decode error: %v\n", err)
	}
}

func TestMaxPacketSize(t *testing.T) {
	bits := 1024
	pk, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		t.Fatal(err)
	}
	size := MaxPacketSize(&pk.PublicKey)
	if size != bits/8 {
		t.Errorf("invlaid max packet size=%v", size)
	}
	payload := size - 2*hashSize - 2 - hashSize - 2
	t.Log(payload)
	if payload != MaxPacketPayloadSize(&pk.PublicKey) {
		t.Errorf("invlaid payload size=%v", payload)
	}
}
