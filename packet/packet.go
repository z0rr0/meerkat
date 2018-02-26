// Copyright 2018 Alexander Zaytsev <thebestzorro@yandex.ru>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// Package main implements client/server common part - packet settings/methods.
package packet

import (
	"crypto/rsa"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const (
	// hashSize is SHA256 hash bytes size.
	hashSize = 32
	// InterruptPrefix is constant prefix of interrupt signal
	InterruptPrefix = "interrupt signal"
)

// Packet is main packet structure.
type Packet struct {
	Name    string
	Payload []byte
}

// MaxPacketSize is max total packet size/
func MaxPacketSize(publicKey *rsa.PublicKey) int {
	return publicKey.N.BitLen() / 8
}

// MaxPacketPayloadSize calculates max UDP packet size.
func MaxPacketPayloadSize(publicKey *rsa.PublicKey) int {
	return MaxPacketSize(publicKey) - 2*hashSize - 2
}

// Interrupt catches custom signals.
func Interrupt(ec chan error) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	ec <- fmt.Errorf("%v %v", InterruptPrefix, <-c)
}
