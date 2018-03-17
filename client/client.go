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

// Package main implements client part of Meerkat project.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/z0rr0/meerkat/conf"
	"github.com/z0rr0/meerkat/packet"
)

const (
	// Name is a program name.
	Name = "Meerkat client"
	//// interruptPrefix is constant prefix of interrupt signal
	//interruptPrefix = "interrupt signal"
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
	defer func() {
		if r := recover(); r != nil {
			loggerError.Printf("Unexpected failed\n%v\n", r)
			os.Exit(1)
		}
	}()
	version := flag.Bool("version", false, "only print version")
	config := flag.String("config", "meerkat_client.json", "client configuration file")
	flag.Parse()

	if *version {
		fmt.Printf("%v: %v %v %v %v\n", Name, Version, Revision, GoVersion, Date)
		return
	}
	cfg := &Config{}
	err := conf.ReadFile(cfg, *config)
	if err != nil {
		loggerError.Fatalln(err)
	}
	loggerInfo.Printf("configuration is read\nServer %v:%v\n", cfg.Server.Host, cfg.Server.Port)

	errChan := make(chan error)
	defer close(errChan)

	go packet.Interrupt(errChan)
	go Run(cfg, errChan)

	loggerError.Println(<-errChan)
}
