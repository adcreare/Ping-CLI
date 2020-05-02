package main

import (
	"time"
	"fmt"
	"net"
	"os"
	"log"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

const (
	ProtocolICMP = 1
	ProtocolIPv6ICMP = 58
)

func ping(address string) (*net.IPAddr, time.Duration, error) {

	// if the input is a DNS, resolve, then get the real address

	dst,error := net.ResolveIPAddr("ip",address)
		if error != nil {
			log.Fatal(error)
	}

	// if we are dealing with ipv4...

	if dst.IP.To4() != nil {

		// create a listener

		c,error := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
		if error != nil {
			return nil,0,error
		}
		defer c.Close()
		
		// create a message

		m := icmp.Message {
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo {
				ID: os.Getpid() & 0xffff, Seq: 1,
				Data: []byte(""),
			},
		}

		b, error := m.Marshal(nil)
		if error != nil{
			log.Fatal(error)
		}

		// start waiting for replies to messages, and tracking the RTT

		start := time.Now()
		n, error := c.WriteTo(b,dst)
		if error != nil {
			log.Fatal(error)
		}

		response := make([]byte, 1500)
		error = c.SetReadDeadline(time.Now().Add(10 * time.Second))
		if error != nil {
			return dst, 0, error
		}
		n, peer, error := c.ReadFrom(response)
		if error != nil {
			return dst,0,error
		}
		RTT := time.Since(start)

		rm, error := icmp.ParseMessage(ProtocolICMP, response[:n])
		if error != nil {
			return dst,0,error
		}

		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			return dst, RTT, nil
		default:
			return dst, 0, fmt.Errorf("got %+v from %v",rm,peer)
		}
	} else { // else, we are dealing with ipv6...
		c,error := icmp.ListenPacket("ip6:58", "::")
		if error != nil {
			return nil,0,error
		}
		defer c.Close()
		
		m := icmp.Message {
			Type: ipv6.ICMPTypeEchoRequest, Code: 0,
			Body: &icmp.Echo {
				ID: os.Getpid() & 0xffff, Seq: 1,
				Data: []byte(""),
			},
		}

		b, error := m.Marshal(nil)
		if error != nil{
			log.Fatal(error)
		}

		start := time.Now()
		n, error := c.WriteTo(b,dst)
		if error != nil {
			log.Fatal(error)
		}

		response := make([]byte, 1500)
		error = c.SetReadDeadline(time.Now().Add(10 * time.Second))
		if error != nil {
			return dst,0,error
		}
		n, peer, error := c.ReadFrom(response)
		if error != nil {
			return dst,0,error
		}
		RTT := time.Since(start)

		rm, error := icmp.ParseMessage(ProtocolIPv6ICMP, response[:n])
		if error != nil {
			return dst,0,error
		}

		switch rm.Type {
		case ipv6.ICMPTypeEchoReply:
			return dst, RTT, nil
		default:
			return dst, 0, fmt.Errorf("got %+v from %v",rm,peer)
	}
}


}

func main() {
	sentCount := 0
	receivedCount := 0
	address := os.Args[1] // take in the first arg input in the command line as our address

	// start an infinite loop of pings

	for {
		sentCount++ // each time a loop is initiated, a ping is sent
		ping := func(address string) {
			dst, rtt, error := ping(address)
			if error != nil {
				log.Printf("Ping: * (*), RTT: * \n")
			} else {
				log.Printf("Ping: %s (%s), RTT: %s\n",address, dst, rtt)
				receivedCount++ // track the amount of pings received
			}
		}
		ping(address) // run the ping function
		pl := (sentCount-receivedCount)/sentCount // calculate packet loss percentage
		log.Printf("Packet Loss: %v %% (%v packets lost) \n",pl,sentCount-receivedCount)
		time.Sleep(2 * time.Second) // 2 second delay time between loops
	}
}
