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
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const (
	// KeySize is RSA public key size.
	KeySize = 512 // 4096 / 8
	// MaxPacketSize is max UDP packet size.
	MaxPacketSize = 446 // 446 = KeySize - 2*(hash size) - 2 = 512 - 64 - 2
	// InterruptPrefix is constant prefix of interrupt signal
	InterruptPrefix = "interrupt signal"
)

// Packet is main packet structure.
type Packet struct {
	Name    string
	Payload []byte
}

// Interrupt catches custom signals.
func Interrupt(ec chan error) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	ec <- fmt.Errorf("%v %v", InterruptPrefix, <-c)
}
