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
)

const version string = "0.1"
const maxBufferSize = 1024

var (
        address         = flag.String("a", "0.0.0.0:8080", "Listen IP address")
        debug           = flag.Bool("debug", false, "Enable verbose output")
        release     = flag.Bool("v", false, "Show version info")
        ifaces          = flag.Bool("l", false, "List available interfaces")
)

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

        for _, i := range interfaces {
                addrs, err := i.Addrs()
                if err != nil {continue}

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

        go startCollector()

        startReceiver()
}

func startCollector() {
        // Open output pcap file and write header
        f, _ := os.Create("test.pcap")
        w := pcapgo.NewWriter(f)
        w.WriteFileHeader(65536, layers.LinkTypeEthernet)
        defer f.Close()

        // Open the device for capturing
        handle, err = pcap.OpenLive("ens18", 65536, false, pcap.BlockForever)
        if err != nil {
                fmt.Printf("Error opening device %s: %v", "ens18", err)
                os.Exit(1)
        }
        defer handle.Close()

        // Set Filter
        var filter string = "udp and port 8080"
        err = handle.SetBPFFilter(filter)
        if err != nil {log.Fatal(err)}

        fmt.Printf("Starting capture.\n")

        // Start processing packets
        packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
        for packet := range packetSource.Packets() {
                // Process packet here
                fmt.Println(packet)
                w.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
                packetCount++

                // Only capture 100 and then stop
                //if packetCount > 100 {
                //      break
                //}
        }
}

func startReceiver() {
        pc, err := net.ListenPacket("udp", *address)
        if err != nil {log.Fatal(err)}
        defer pc.Close()

        if *debug {
                fmt.Printf("Server started. Listening for UDP packets on %s\n", *address)
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
