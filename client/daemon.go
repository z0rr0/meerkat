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
// Package main implements client part of Meerkat project.
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"os/exec"
	"sync"
	"time"

	"github.com/z0rr0/meerkat/packet"
)

var (
	workersMap = map[string]func(*Service, int, chan<- *packet.Packet, *sync.WaitGroup){
		"command": workerCommand,
		//"memory"
		//"cpu"
	}
)

// workerCommand is a common service worker.
func workerCommand(s *Service, packetSize int, co chan<- *packet.Packet, wg *sync.WaitGroup) {
	var err error
	defer wg.Done()

	buf := make([]byte, packetSize)
	p := &packet.Packet{Name: s.Name, Payload: make([]byte, packetSize)}

	loggerInfo.Printf("run worker [%v], period=%v seconds\n", s.Name, s.Period)
	d := time.Duration(s.Period) * time.Second
	timer := time.NewTimer(d)
	defer timer.Stop()

	for range timer.C {
		buf, err = exec.Command(s.Exec, s.Args...).Output()
		if err != nil {
			loggerError.Printf("worker [%v] [ignore=%v], error: %v\n", s.Name, s.IgnoreErrors, err)
			if !s.IgnoreErrors {
				// error without ignoring, exit
				return
			}
		}
		if l := len(buf); l > packetSize {
			loggerError.Printf("worker [%v], too match packet %v bytes", s.Name, l)
		} else {
			loggerInfo.Printf("worker [%v]: %v bytes\n", s.Name, l)
			copy(p.Payload, buf[0:l])
			co <- p
		}
		timer.Reset(d)
	}
}

// consume handles command outputs.
func consume(s *Server, co <-chan *packet.Packet) {
	for out := range co {
		// send to server
		encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, s.publicKey, out.Payload, nil)
		if err != nil {
			loggerError.Printf("error encrypted, worker [%v] - %v bytes: %v\n", out.Name, len(out.Payload), err)
		} else {
			loggerInfo.Printf("handle worker [%v] message [%v]: \n%v\n", out.Name, len(out.Payload), string(out.Payload))
			err = s.send(encrypted)
			if err != nil {
				loggerError.Printf("error during message sending: %v\n", err)
			}
		}
	}
}

// Run starts main services.
func Run(cfg *Config, ec chan error) {
	var wg sync.WaitGroup
	if l := len(cfg.Services); l == 0 {
		ec <- errors.New("no services for running")
		return
	} else {
		wg.Add(l)
	}

	co := make(chan *packet.Packet)
	defer close(co) // only if no working services
	go consume(&cfg.Server, co)

	for i, s := range cfg.Services {
		if worker, ok := workersMap[s.Type]; ok {
			go worker(&cfg.Services[i], packet.MaxPacketPayloadSize(cfg.Server.publicKey), co, &wg)
		} else {
			loggerError.Printf("unknown service [%v] type: '%v'\n", s.Name, s.Type)
			wg.Done()
		}
	}
	wg.Wait()
	ec <- nil
}
