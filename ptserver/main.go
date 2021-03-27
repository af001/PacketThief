package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

const version string = "0.1"
const maxBufferSize = 1024

var (
	address = flag.String("a", "0.0.0.0", "Listen IP address")
	debug	= flag.Bool("debug", false, "Enable verbose output")
	release     = flag.Bool("v", false, "Show version info")
)

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: ptserver [ ... ]\n\nParameters:")
		flag.PrintDefaults()
	}
}

func printVersion() {
	fmt.Println("Packet Thief Server")
	fmt.Printf("Version: %s\n", version)
}

func main() {
	flag.Parse()

	if *release {
		printVersion()
		os.Exit(0)
	}

	pc, err := net.ListenPacket("udp", *address)
	if err != nil {log.Fatal(err)}
	defer pc.Close()

	fmt.Printf("Server started. Listening for UDP packets on %s", *address)

	doneChan := make(chan error, 1)
	buffer := make([]byte, maxBufferSize)

	go func() {
		for {
			n, addr, err := pc.ReadFrom(buffer)
			if err != nil {
				doneChan <- err
				return
			}

			if *debug {
				fmt.Printf("Received %d bytes from %s\n", n, addr.String())
			}
		}
	}()
}
