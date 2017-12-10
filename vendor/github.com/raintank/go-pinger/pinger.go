/*
	Package pinger provides a service for sending ICMP echo requests to numerous hosts in parallel.
	The results from each "Ping" issued can used to calculate min,max,avg and stdev latency and % loss.

	A process should only create 1 pinger service which can be shared across multiple goroutines.

	see github.com/raintank/go-pinger/ping-all for an example of how to use the service.
*/
package pinger

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

/*
	PingStats are the results from sending ICMP echo requests to a host.
	The stats include the number of packets sent and received and the latency
	for every packet received.
*/
type PingStats struct {
	Latency  []time.Duration
	Sent     int
	Received int
}

/*
	Pinger represents a service that can be used for sending IMCP pings to hosts.
*/
type Pinger struct {
	inFlight    map[string]*EchoRequest
	v4Conn      net.PacketConn
	v6Conn      net.PacketConn
	packetChan  chan *EchoResponse
	requestChan chan *EchoRequest
	Counter     int
	proto       string
	processWg   *sync.WaitGroup
	Debug       bool
	shutdown    bool

	sync.RWMutex
}

/*
	Creates a new Pinger service.  Accepts the IP protocol to use "ipv4", "ipv6" or "all" and
	the number of packets to buffer in the request and response packet channels.
	The pinger instance will immediately start listening on the raw sockets (ipv4:icmp, ipv6:ipv6-icmp or both).
*/
func NewPinger(protocol string, bufferSize int) (*Pinger, error) {
	rand.Seed(time.Now().UnixNano())

	p := &Pinger{
		inFlight:    make(map[string]*EchoRequest),
		Counter:     rand.Intn(0xffff),
		proto:       protocol,
		packetChan:  make(chan *EchoResponse, bufferSize),
		requestChan: make(chan *EchoRequest, bufferSize),
		processWg:   new(sync.WaitGroup),
	}
	var err error
	switch protocol {
	case "ipv4":
		p.v4Conn, err = net.ListenPacket("ip4:icmp", "0.0.0.0")
		if err != nil {
			return nil, err
		}
	case "ipv6":
		p.v6Conn, err = net.ListenPacket("ip6:ipv6-icmp", "::")
		if err != nil {
			return nil, err
		}
	case "all":
		p.v4Conn, err = net.ListenPacket("ip4:icmp", "0.0.0.0")
		if err != nil {
			return nil, err
		}
		p.v6Conn, err = net.ListenPacket("ip6:ipv6-icmp", "::")
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid protocol, must be ipv4 or ipv6")
	}

	return p, nil

}

/*
	Start launches goroutines for processing packets received on the raw sockets
*/
func (p *Pinger) Start() {
	if p.proto == "all" || p.proto == "ipv4" {
		go p.v4PacketReader()
	}
	if p.proto == "all" || p.proto == "ipv6" {
		go p.v6PacketReader()
	}
	p.processWg.Add(1)
	go p.processPkt()
}

/*
	Stop shuts down the pinger service and closes the raw sockets. This method will block
	until all spawned goroutines have ended.
*/
func (p *Pinger) Stop() {
	p.Lock()
	p.shutdown = true
	p.Unlock()
	if p.v4Conn != nil {
		p.v4Conn.Close()
	}
	if p.v6Conn != nil {
		p.v6Conn.Close()
	}
	close(p.packetChan)
	p.processWg.Wait()
}

/*
 Send <count> icmp echo rquests to <address> and don't wait longer then <timeout> for a response.
 An error will be returned if the EchoRequests cant be sent.
 This call will block until all icmp EchoResponses are received or timeout is reached. It is safe
 to call this method concurrently.
*/
func (p *Pinger) Ping(address net.IP, count int, timeout time.Duration) (*PingStats, error) {
	p.Lock()
	if p.shutdown {
		p.Unlock()
		return nil, fmt.Errorf("Pinger service is shutdown.")
	}
	p.Counter++
	if p.Counter > 65535 {
		p.Counter = 0
	}
	supportedProto := p.proto
	counter := p.Counter
	p.Unlock()

	var proto icmp.Type
	proto = ipv4.ICMPTypeEcho
	if address.To4() == nil {
		proto = ipv6.ICMPTypeEchoRequest
	}

	if proto == ipv4.ICMPTypeEcho && supportedProto == "ipv6" {
		return nil, fmt.Errorf("This pinger instances does not support ipv4")
	}
	if proto == ipv6.ICMPTypeEchoRequest && supportedProto == "ipv4" {
		return nil, fmt.Errorf("This pinger instances does not support ipv6")
	}

	pingTest := make([]*EchoRequest, count)
	wg := new(sync.WaitGroup)

	p.Lock()
	for i := 0; i < count; i++ {
		wg.Add(1)
		pkt := icmp.Message{
			Type: proto,
			Code: 0,
			Body: &icmp.Echo{
				ID:   counter,
				Seq:  i,
				Data: []byte("raintank/go-pinger"),
			},
		}

		req := NewEchoRequest(pkt, address, wg)
		pingTest[i] = req

		// record our packet in the inFlight queue
		p.inFlight[req.ID] = req
	}
	p.Unlock()

	for _, req := range pingTest {
		if p.Debug {
			log.Printf("go-pinger: sending packet. Peer %s, Id: %d, Seq: %d, Sent: %s", address.String(), req.Body.(*icmp.Echo).ID, req.Body.(*icmp.Echo).Seq, time.Now().String())
		}
		err := p.Send(req)
		if err != nil {
			// cleanup requests from inFlightQueue
			p.Lock()
			for _, r := range pingTest {
				delete(p.inFlight, r.ID)
			}
			p.Unlock()
			return nil, err
		}
	}

	// wait for all packets to be received
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// wait for all packets to be recieved or for timeout.
	select {
	case <-done:
		if p.Debug {
			log.Printf("go-pinger: all pings set to %s were received", address.String())
		}
	case <-time.After(timeout):
		if p.Debug {
			log.Printf("go-pinger: timeout reached sending to %s", address.String())
		}
		p.Lock()
		for _, req := range pingTest {
			_, ok := p.inFlight[req.ID]
			if ok {
				delete(p.inFlight, req.ID)
				// we never received a response.  To prevent leaking goroutines we need to
				// ensure our waitgroup reaches 0.
				req.WaitGroup.Done()
			}
		}
		p.Unlock()
	}

	// calculate our timing stats.
	stats := new(PingStats)
	for _, req := range pingTest {
		stats.Sent++
		if !req.Received.IsZero() {
			stats.Received++
			stats.Latency = append(stats.Latency, req.Received.Sub(req.Sent))
		} else {

		}
	}
	return stats, nil
}
