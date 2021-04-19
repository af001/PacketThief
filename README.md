# PacketThief
Port mirroring tool written in Go to send traffic to an external server

```bash
git clone https://github.com/af001/PacketThief.git
```

### Example static cross-compilation example for ARM (ptclient) and X86-64 (ptserver) 
```bash
# Install dependencies for libpcap cross-compilation
apt update && apt -y install flex bison

# Get libpcap
wget https://www.tcpdump.org/release/libpcap-1.10.0.tar.gz
tar xzvf libpcap-1.10.0.tar.gz
cd libpcap-1.10.0
```

### Build
This server will lisen on a port and write data to a pcap file. 
```bash
cd PacketThief/ptserver

# Download deps
go get -u github.com/google/gopacket

# Build 
CC=gcc GOOS=linux GOARCH=amd64 go build -v -o ptserver-amd64 -ldflags '-w -extldflags "-static"' .
strip ptserver-amd64

# Start server
./ptserver-amd64 -i ens18 -p 8000 -w capture.pcap
```

The client will capture traffic and send it to the server.
```bash
# Download deps
go get -u github.com/erikdubbelboer/gspt
go get -u github.com/google/gopacket

# Configure and compile for a specific architecture
apt install -y gcc-multilib-arm-linux-gnueabi
export CC=arm-linux-gnueabi-gcc
export CFLAGS='-Os'
cd libpcap-1.10.0
./configure --host=arm-linux-gnueabi --with-pcap=linux

# Build
cd PacketThief/ptclient
CC=arm-linux-gnueabi-gcc GOOS=linux GOARCH=arm CGO_ENABLED=1 CGO_LDFLAGS="-L /root/libpcap-1.10.0" go build -v -o ptclient-arm -ldflags '-w -extldflags "-static"' .
arm-linux-gnueabi-strip ptclient-arm

# Start client
./ptclient-arm -r 192.168.10.10:8000 -i any -p 1194 
```

### Usage

**Server**
```bash
Usage: ptserver [ ... ]

Parameters:
  -a string
    	The host address to capture packets from
  -debug
    	Enable verbose output
  -i string
    	Interface to listen on (default "any")
  -l	List available interfaces
  -p int
    	Port to listen on
  -s int
    	Capture snap length (default 65536)
  -t string
    	Protocol type to capture (tcp|udp) (default "udp")
  -v	Show version info
  -w string
    	Pcap filename (default "dump.pcap")
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

#### Issues
* When listening on *any* interface, the capture reconstruction may be incorrect. This has to do with the decoding Linux Cooked Capture layer. [Info](https://github.com/google/gopacket/issues/37) 

#### References
[udp-mirror](https://github.com/czerwonk/udp-mirror) by czerwonk
