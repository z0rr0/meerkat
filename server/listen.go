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
// Package main implements server part of Meerkat project.

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"net"
	"strings"
	"sync"

	"github.com/z0rr0/meerkat/packet"
)

func receive(privateKey *rsa.PrivateKey, msg []byte) error {
	text, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, msg, nil)
	if err != nil {
		return err
	}
	loggerInfo.Printf("receive data\n%v\n", string(text))
	return nil
}

// listen reads data from UDP socket
func listen(udpConn *net.UDPConn, privateKey *rsa.PrivateKey, wg *sync.WaitGroup, stop chan bool) {
	wg.Add(1)
	defer wg.Done()

	bc := make(chan []byte)
	go func() {
		var buf [packet.KeySize]byte
		for {
			n, addr, err := udpConn.ReadFromUDP(buf[:])
			if err != nil {
				if msg := err.Error(); strings.Contains(msg, "use of closed network connection") {
					loggerInfo.Println(err)
					close(bc)
					return
				}
				loggerError.Println(err)
			}
			loggerInfo.Printf("read %v bytes from %v\n", n, addr)
			bc <- buf[:n]
		}
	}()

	for {
		select {
		case <-stop:
			return
		case b := <-bc:
			// handled incoming data
			if err := receive(privateKey, b); err != nil {
				loggerError.Printf("error during message decoding: %v\n", err)
			}
		}
	}
}
