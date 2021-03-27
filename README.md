# PacketThief
Port mirroring tool written in Go to send traffic to an external server

### Build
This server will lisen on a port and write data to a pcap file. 
```bash
cd ptserver

# Download deps
go get -t github.com/google/gopacket

# Build 
go build .
```

The client will capture traffic and send it to the server.
```bash
cd ptclient

# Download deps
go get -t github.com/erikdubbelboer/gspt
go get -t github.com/google/gopacket

# Build
go build .
```

### Usage

**Server**
```bash
Usage: ptclient [ ... ]

Parameters:
  -a string
    	The host address to capture packets from
  -debug
    	Debug mode
  -i string
    	Interface to get packets from (default "any")
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
