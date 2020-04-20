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
var icmpType icmp.Type

func showStats(host string) {
	loss := float64(pktsSent-pktsRcvd) / float64(pktsSent) * 100
	fmt.Printf("\n--- %v ping results ---\n", host)
	fmt.Printf("\n%d packets transmitted, %d received, %.2f%% packet loss, time %dms\n",
		pktsSent, pktsRcvd, loss, totalTime.Milliseconds())

	os.Exit(0)
}

func ping(addr string, ipaddr *net.IPAddr, ttl int, packetSize int, quiet bool) {

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
		if !quiet {
			fmt.Printf("\n%d bytes from %s: icmp_seq=%d time=%vms loss=%.2f%%",
				n, ipaddr, sequence, duration.Milliseconds(), loss)
		}
	default:
		if !quiet {
			fmt.Println("TTL Exceeded, Provided TTL is", ttl)
		}
	}
}

func main() {

	host := "cloudflare.com"

	ttl := 64
	flag.IntVar(&ttl, "ttl", 64, "Number of hops before a packet dies")

	packetSize := 8
	flag.IntVar(&packetSize, "s",8 , "Size of a packet (in bytes). Max is 70 bytes")

	quiet := false
	flag.BoolVar(&quiet, "q", false, "Quiet output. Nothing is printed except start and finish summary lines")

	count := -1
	flag.IntVar(&count, "c", -1, "Stop after sending count ECHO_REQUEST packets")

	interval := 1.0
	flag.Float64Var(&interval, "i", 1, "Wait interval seconds between sending each packet")

	deadline := -1
	flag.IntVar(&deadline, "w", -1, "Specify a timeout, in seconds, before ping exits regardless of how many packets have been sent or received")

	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("No host have been provided... Using cloudflare as the default host")

	} else {
		host = args[0]
	}

	keyboardInterrupt := make(chan os.Signal, 1)
	signal.Notify(keyboardInterrupt, os.Interrupt)

	ipaddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	fmt.Println("Started pinging", host, "at IP address", ipaddr)

	go func() {
		for {
			select {
			case _ = <-keyboardInterrupt:
				showStats(host)
			}
		}
	}()

	if deadline > -1 {
		go func() {
			for {
				select {
				case <-time.After(time.Duration(deadline) * time.Second):
					showStats(host)
				}
			}
		}()
	}

	for i := 0; i < count || count == -1; i++ {
		ping(host, ipaddr, ttl, packetSize, quiet)
		time.Sleep(time.Duration(interval) * time.Second)
	}
	showStats(host)

}
