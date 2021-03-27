# PacketThief
Port mirroring tool written in Go to send traffic to an external server

### Build
This server will lisen on a port and write data to a pcap file. 
```bash
cd ptserver

# Download deps
go get -u github.com/google/gopacket

# Build 
go build .
```

The client will capture traffic and send it to the server.
```bash
cd ptclient

# Download deps
go get -u github.com/erikdubbelboer/gspt
go get -u github.com/google/gopacket

# Build
go build .
```

### Usage

**Server**
```bash
Usage: ptserver [ ... ]

Parameters:
  -a string
    	Listen IP address (default "0.0.0.0:8080")
  -debug
    	Enable verbose output
  -l	List available interfaces
  -v	Show version info
```

**Client**
```bash
Usage: ptclient [ ... ]

Parameters:
  -a string
    	The host address to capture packets from
  -debug
    	Debug mode
  -i string
    	Interface to get packets from (default "any")
  -l	List available interfaces
  -n string
    	Set a custom process name (default "iomemd")
  -p int
    	Port to get packets from
  -r string
    	Comma separated list of receivers (ip:port)
  -s int
    	SnapLen for packet capture (default 65536)
  -t string
    	Protocol to capture (udp|tcp) (default "udp")
  -v	Show version info
```

#### References
[udp-mirror](https://github.com/czerwonk/udp-mirror) by czerwonk
