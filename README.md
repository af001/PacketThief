# PacketThief
Port mirroring tool written in Go to send traffic to an external server

#### Build Server
This server will lisen on a port and write data to a pcap file. 
```bash
cd ptserver

# Download deps
go get -t github.com/google/gopacket

# Build 
go build .
```

The client will capture traffic and send it to the server.
```
cd ptclient

# Download deps
go get -t github.com/erikdubbelboer/gspt
go get -t github.com/google/gopacket

# Build
go build .
```


#### References
[udp-mirror](https://github.com/czerwonk/udp-mirror) by czerwonk
