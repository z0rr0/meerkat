package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
)

const (
	// Name is a program name.
	Name = "Meerkat"
	// interruptPrefix is constant prefix of interrupt signal
	interruptPrefix = "interrupt signal"
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

// interrupt catches custom signals.
func interrupt(ec chan error) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	ec <- fmt.Errorf("%v %v", interruptPrefix, <-c)
}

// listen reads data from UDP socket
func listen(udpConn *net.UDPConn, wg *sync.WaitGroup, stop chan bool) {
	wg.Add(1)
	defer wg.Done()

	bc := make(chan []byte)

	go func() {
		var buf [4096]byte
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
			loggerInfo.Printf("read from %v\n", addr)
			bc <- buf[:n]
		}
	}()

	for {
		select {
		case <-stop:
			return
		case b := <-bc:
			// handled incoming data
			fmt.Printf("data:\n%v\n", b)
		}
	}
}

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
	cfg, err := Configuration(*config)
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

	go interrupt(errChan)
	go listen(udpConn, &wg, stopChan)

	// wait error or valid interrupt
	err = <-errChan
	// stop upd socket read
	close(stopChan)
	// wait graceful stop
	wg.Wait()

	loggerInfo.Println("gracefully stopped")
}
