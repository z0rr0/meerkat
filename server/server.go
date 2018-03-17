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

// Package main implements server part of Meerkat project.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"sync"

	"github.com/z0rr0/meerkat/conf"
	"github.com/z0rr0/meerkat/packet"
)

const (
	// Name is a program name.
	Name = "Meerkat"
)

var (
	// Version is a version from GIT tags
	Version = "0.0.0"
	// Revision is GIT revision number
	Revision = "git:000000"
	// Date is build date
	Date = "2016-01-01_01:01:01UTC"
	// GoVersion is runtime Go language version
	GoVersion = runtime.Version()

	// loggerError is a logger for error messages
	loggerError = log.New(os.Stderr, fmt.Sprintf("%v [ERROR]: ", Name), log.Ldate|log.Lmicroseconds|log.Lshortfile)
	// loggerInfo is a logger for debug/info messages
	loggerInfo = log.New(os.Stdout, fmt.Sprintf("%v [INFO]: ", Name), log.Ldate|log.Ltime|log.Lshortfile)
)

func main() {
	var wg sync.WaitGroup
	defer func() {
		if r := recover(); r != nil {
			loggerError.Printf("Unexpected failed\n%v\n", r)
			os.Exit(1)
		}
	}()
	version := flag.Bool("version", false, "only print version")
	genkeys := flag.Int("genkeys", 0, "generate RSA keys pair")
	config := flag.String("config", "meerkat.json", "configuration file")
	flag.Parse()

	if *version {
		fmt.Printf("%v: %v %v %v %v\n", Name, Version, Revision, GoVersion, Date)
		return
	}
	if *genkeys > 0 {
		GenKeys(*genkeys, loggerError)
		return
	}
	cfg := &Config{}
	err := conf.ReadFile(cfg, *config)
	if err != nil {
		loggerError.Fatalln(err)
	}
	loggerInfo.Printf("configuration is read\n%v:%v\n", cfg.Server.Host, cfg.Server.Port)

	udpConn, err := net.ListenUDP("udp", cfg.Server.UDPAddr())
	if err != nil {
		loggerError.Fatalln(err)
	}
	defer udpConn.Close()

	errChan := make(chan error)
	stopChan := make(chan bool)
	defer close(errChan)

	go packet.Interrupt(errChan)
	go listen(udpConn, cfg.Server.privateKey, &wg, stopChan)

	// wait error or valid interrupt
	err = <-errChan
	// stop upd socket read
	close(stopChan)
	// wait graceful stop
	wg.Wait()

	loggerInfo.Println("gracefully stopped")
}
