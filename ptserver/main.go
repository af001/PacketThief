package main

import (
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"log"
	"net"
	"os"
	"time"
)

const version string = "0.1"
const maxBufferSize = 1024

var (
	device          = flag.String("i", "any", "Interface to listen on")
	port            = flag.Int("p", 0, "Port to listen on")
	proto			= flag.String("t", "udp", "Protocol type to capture (tcp|udp)")
	target			= flag.String("a", "", "The host address to capture packets from")
	name			= flag.String("w", "dump.pcap", "Pcap filename")
	snaplen			= flag.Int("s", 65536, "Capture snap length")
	debug           = flag.Bool("debug", false, "Enable verbose output")
	release     	= flag.Bool("v", false, "Show version info")
	ifaces          = flag.Bool("l", false, "List available interfaces")
)

var cDecodeLayer = gopacket.RegisterLayerType(12345, gopacket.LayerTypeMetadata{Name: "cDecodeLayer", Decoder: gopacket.DecodeFunc(layerDecoder)})

type cLayer struct {
	cHeader []byte
	payload []byte
}

func layerDecoder(data []byte, p gopacket.PacketBuilder) error {
	p.AddLayer(&cLayer{data[:16], data[16:]})
	return p.NextDecoder(layers.LayerTypeIPv4)
}

func (m cLayer) LayerType() gopacket.LayerType { return cDecodeLayer }
func (m cLayer) LayerContents() []byte { return m.cHeader }
func (m cLayer) LayerPayload() []byte { return m.payload }

var (
	err         error
	handle      *pcap.Handle
	packetCount int = 0
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

func showInterfaces() {
	interfaces, err := net.Interfaces()
	if err != nil {log.Fatal(err)}

	fmt.Println("Interfaces: \n\nDevice\t  Address")

	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {continue}

		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPAddr:
				fmt.Printf("%v\t: %s\n", i.Name, v)

			case *net.IPNet:
				fmt.Printf("%v\t: %s\n", i.Name, v)
			}
		}
	}
	fmt.Printf("\n")
}

func getInterfaceIpv4Addr(interfaceName string) (addr string, err error) {
	var (
		ief      *net.Interface
		addrs    []net.Addr
		ipv4Addr net.IP
	)

        if interfaceName == "any" {
                return "0.0.0.0", nil
        }

	if ief, err = net.InterfaceByName(interfaceName); err != nil {
		return
	}

	if addrs, err = ief.Addrs(); err != nil {
		return
	}

	for _, addr := range addrs {
		if ipv4Addr = addr.(*net.IPNet).IP.To4(); ipv4Addr != nil {
			break
		}
	}

	if ipv4Addr == nil {
		ipv4Addr = net.IP("0.0.0.0")
	}

	return ipv4Addr.String(), nil
}

func main() {
	flag.Parse()

	if *release {
		printVersion()
		os.Exit(0)
	}

	if *ifaces {
		showInterfaces()
		os.Exit(0)
	}

	protoChoices := map[string]bool{"udp": true, "tcp": true}
	if _, validChoice := protoChoices[*proto]; !validChoice {
		fmt.Println("Invalid protocol")
		os.Exit(1)
	}

	go startCollector()

	startReceiver()
}

func buildBPF() string {
	var filter string
	if *port == 0 && len(*target) == 0{
		fmt.Printf("Missing port or host filter\n")
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

func startCollector() {
	// Open output pcap file and write header
	f, _ := os.Create(*name)
	w := pcapgo.NewWriter(f)
	_ = w.WriteFileHeader(65536, layers.LinkTypeEthernet)
	defer f.Close()

	// Open the device for capturing
	handle, err = pcap.OpenLive(*device, int32(*snaplen), false, pcap.BlockForever)
	if err != nil {
		fmt.Printf("Error opening device %s: %v", *device, err)
		os.Exit(1)
	}
	defer handle.Close()

	// Set Filter
	filter := buildBPF()
	err = handle.SetBPFFilter(filter)
	if err != nil {log.Fatal(err)}

	if *debug {
		fmt.Printf("Starting capture.\n")
	}

	// Start processing packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		udpLayer := packet.Layer(layers.LayerTypeUDP)
		if udpLayer != nil {
			udp, _ := udpLayer.(*layers.UDP)
			pkt := gopacket.NewPacket(udp.Payload, cDecodeLayer, gopacket.Lazy)
			err = w.WritePacket(gopacket.CaptureInfo{Timestamp: time.Now(), Length: len(pkt.Data()), CaptureLength: len(pkt.Data()), InterfaceIndex: 0}, pkt.Data())
			if err != nil {
				fmt.Println(err)
			}
			packetCount++
		}
	}
}

func startReceiver() {
	listen, err := getInterfaceIpv4Addr(*device)
	if err != nil {
		fmt.Println("Invalid interface")
		os.Exit(1)
	}

	pc, err := net.ListenPacket("udp", fmt.Sprintf("%s:%d", listen, *port))
	if err != nil {log.Fatal(err)}
	defer pc.Close()

	if *debug {
		fmt.Printf("Server started. Listening for %s packets on %s:%d\n", *proto, listen, *port)
	}

	doneChan := make(chan error, 1)
	buffer := make([]byte, maxBufferSize)

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
}
