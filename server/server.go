package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
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
	_, err := Configuration(*config)
	if err != nil {
		loggerError.Fatalln(err)
	}
	fmt.Println("ok")
}
