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
	"errors"
	"os/exec"
	"sync"
	"time"

	"github.com/z0rr0/meerkat/packet"
)

// output is command result
type output struct {
	s *Service
	b []byte
}

// worker is a common service worker.
func worker(cfg *Config, s *Service, co chan<- *output, wg *sync.WaitGroup) {
	var err error
	defer wg.Done()

	out := make([]byte, packet.MaxPacketSize)
	o := &output{s: s, b: make([]byte, packet.MaxPacketSize)}

	loggerInfo.Printf("run worker [%v], period=%v seconds\n", s.Name, s.Period)
	d := time.Duration(s.Period) * time.Second
	timer := time.NewTimer(d)
	defer timer.Stop()

	for range timer.C {
		out, err = exec.Command(s.Exec, s.Args...).Output()
		if err != nil {
			loggerError.Printf("worker [%v] [ignore=%v], error: %v\n", s.Name, s.IgnoreErrors, err)
			if !s.IgnoreErrors {
				// error without ignoring, exit
				return
			}
		}
		if l := len(out); l > packet.MaxPacketSize {
			loggerError.Printf("worker [%v], too match packet %v bytes", s.Name, l)
		} else {
			loggerInfo.Printf("worker [%v]: %v bytes\n", s.Name, l)
			copy(o.b, out[0:])
			co <- o
		}
		timer.Reset(d)
	}
}

// consume handles command outputs.
func consume(co <-chan *output) {
	for out := range co {
		// send to server
		loggerInfo.Printf("handle worker [%v]: %v\n", out.s.Name, string(out.b))
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

	co := make(chan *output)
	go consume(co)

	for i := range cfg.Services {
		go worker(cfg, &cfg.Services[i], co, &wg)
	}
	wg.Wait()
	ec <- nil
}
