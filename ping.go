package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

var isIPv4 = false
var pktsRcvd int
var pktsSent int
var sequence int
var totalTime time.Duration

func ping(addr string, ttl int, packetSize int) {
	var icmpType icmp.Type

	ipaddr, err := net.ResolveIPAddr("ip", addr)
	if err != nil {
		fmt.Println(err)
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")

	if ipaddr.IP.To4() != nil {
		if err != nil {
			fmt.Println(err)
		}
		conn.IPv4PacketConn().SetTTL(ttl)
		icmpType = ipv4.ICMPTypeEcho
		isIPv4 = true
	} else {
		conn, err := icmp.ListenPacket("ip6:ipv6-icmp", "::")
		if err != nil {
			fmt.Println(err)
		}
		conn.IPv6PacketConn().SetHopLimit(ttl)
		icmpType = ipv6.ICMPTypeEchoRequest
	}
	defer conn.Close()

	m := icmp.Message{
		Type: icmpType, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: bytes.Repeat([]byte("a"), packetSize),
			// Data: []byte(""),
		},
	}

	sequence++
	pktsSent++

	b, err := m.Marshal(nil)
	if err != nil {
		fmt.Println(err)
	}

	start := time.Now()
	n, err := conn.WriteTo(b, ipaddr)
	if err != nil {
		fmt.Println(err)
	} else if n != len(b) {
		fmt.Println(err)
	}

	reply := make([]byte, 1024)
	err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		fmt.Println(err)
	}

	n, _, _, err = conn.IPv4PacketConn().ReadFrom(reply)

	if isIPv4 == false {
		n, _, _, err = conn.IPv4PacketConn().ReadFrom(reply)
	}

	if err != nil {
		fmt.Println(err)
	}

	duration := time.Since(start)
	totalTime += duration

	var protocol int
	if isIPv4 {
		protocol = 1
	} else {
		protocol = 58
	}

	out, err := icmp.ParseMessage(protocol, reply)
	if err != nil {
		fmt.Println(err)
	}

	switch out.Body.(type) {
	case *icmp.Echo:
		pktsRcvd++
		loss := float64(pktsSent-pktsRcvd) / float64(pktsSent) * 100

		fmt.Printf("\n%d bytes from %s: icmp_seq=%d time=%vms loss=%.2f%%",
			n, ipaddr, sequence, duration.Milliseconds(), loss)

	default:
		fmt.Println("TTL Exceeded, Provided TTL is", ttl)
	}
}

func main() {

	host := "cloudflare.com"

	ttl := 64
	flag.IntVar(&ttl, "ttl", 64, "Number of hops before a packet dies")

	packetSize := 50
	flag.IntVar(&packetSize, "size", 50, "Size of a packet (in bytes)")

	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("No arguments have been provided... Using cloudflare as the default host")

	} else {
		host = args[0]
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	go func() {
		<-sigs
		loss := float64(pktsSent-pktsRcvd) / float64(pktsSent) * 100
		fmt.Printf("\n--- %v ping results ---\n", host)
		fmt.Printf("\n%d packets transmitted, %d received, %.2f%% packet loss, time %dms\n",
			pktsSent, pktsRcvd, loss, totalTime.Milliseconds())

		os.Exit(0)
	}()

	for {
		ping(host, ttl, packetSize)
		time.Sleep(1 * time.Second)
	}

}
