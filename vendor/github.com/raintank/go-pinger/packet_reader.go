package pinger

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// from "golang.org/x/net/internal/iana"
const ProtocolIPv6ICMP = 58 // ICMP for IPv6
const ProtocolICMP = 1      // Internet Control Message

type EchoResponse struct {
	Peer     string
	Id       int
	Seq      int
	Received time.Time
}

func (e *EchoResponse) String() string {
	return fmt.Sprintf("Peer %s, Id: %d, Seq: %d, Recv: %s", e.Peer, e.Id, e.Seq, e.Received)
}

func (p *Pinger) v6PacketReader() {
	runtime.LockOSThread()
	if p.Debug {
		log.Printf("ipv6 listen loop starting.")
	}
	var f ipv6.ICMPFilter
	f.SetAll(true)
	f.Accept(ipv6.ICMPTypeDestinationUnreachable)
	f.Accept(ipv6.ICMPTypePacketTooBig)
	f.Accept(ipv6.ICMPTypeTimeExceeded)
	f.Accept(ipv6.ICMPTypeParameterProblem)
	f.Accept(ipv6.ICMPTypeEchoReply)
	c := ipv6.NewPacketConn(p.v6Conn)
	if err := c.SetICMPFilter(&f); err != nil {
		panic(err)
	}

	rb := make([]byte, 1500)

	var data []byte
	ipconn, ok := p.v6Conn.(*net.IPConn)
	if !ok {
		panic("connection is not IPConn")
	}
	file, err := ipconn.File()
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	fd := file.Fd()

	var pktTime time.Time
	recvTime := syscall.Timeval{}
	for {
		n, peer, err := p.v6Conn.ReadFrom(rb)
		if err != nil {
			p.RLock()
			if !p.shutdown {
				log.Printf("go-pinger: failed to read from ipv6 socket. %s", err)
			}
			p.RUnlock()
			break
		}
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.SIOCGSTAMP), uintptr(unsafe.Pointer(&recvTime)))
		err = nil
		if errno != 0 {
			err = errno
		}
		if err == nil {
			pktTime = time.Unix(0, recvTime.Nano())
		} else {
			pktTime = time.Now()
		}

		rm, err := icmp.ParseMessage(ProtocolIPv6ICMP, rb[:n])
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if rm.Type == ipv6.ICMPTypeEchoReply {
			data = rm.Body.(*icmp.Echo).Data
			if len(data) < 9 {
				log.Printf("go-pinger: invalid data payload from %s. Expected at least 9bytes got %d", peer.String(), len(data))
				continue
			}
			if p.Debug {
				log.Printf("go-pinger: recieved pkt. Peer %s, Id: %d, Seq: %d, Recv: %s\n", peer.String(), rm.Body.(*icmp.Echo).ID, rm.Body.(*icmp.Echo).Seq, pktTime.String())
			}

			// this goroutine needs to read packets from the network as fast as possible so we can get accurate timing information.
			// if the packetChan blocks, latency measurements could start to grow.  If we changed this to a non-blocking write on the
			// channel, then packets that could not be written would appear as packet loss, which is probably worse then higher latency.
			p.packetChan <- &EchoResponse{
				Peer:     peer.String(),
				Seq:      rm.Body.(*icmp.Echo).Seq,
				Id:       rm.Body.(*icmp.Echo).ID,
				Received: pktTime,
			}
		}
	}
	if p.Debug {
		log.Printf("ipv6 listen loop ended.")
	}
}

func (p *Pinger) v4PacketReader() {
	runtime.LockOSThread()
	if p.Debug {
		log.Printf("ipv4 listen loop starting.")
	}

	rb := make([]byte, 1500)

	var data []byte
	ipconn, ok := p.v4Conn.(*net.IPConn)
	if !ok {
		panic("connection is not IPConn")
	}
	file, err := ipconn.File()
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	fd := file.Fd()

	var pktTime time.Time
	recvTime := syscall.Timeval{}
	for {
		n, peer, err := p.v4Conn.ReadFrom(rb)
		if err != nil {
			p.RLock()
			if !p.shutdown {
				log.Printf("go-pinger: failed to read from ipv4 socket. %s", err)
			}
			p.RUnlock()
			break
		}
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.SIOCGSTAMP), uintptr(unsafe.Pointer(&recvTime)))
		err = nil
		if errno != 0 {
			err = errno
		}
		if err == nil {
			pktTime = time.Unix(0, recvTime.Nano())
		} else {
			pktTime = time.Now()
		}

		rm, err := icmp.ParseMessage(ProtocolICMP, rb[:n])
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if rm.Type == ipv4.ICMPTypeEchoReply {
			data = rm.Body.(*icmp.Echo).Data
			if len(data) < 9 {
				log.Printf("go-pinger: invalid data payload from %s. Expected at least 9bytes got %d", peer.String(), len(data))
				continue
			}
			if p.Debug {
				log.Printf("go-pinger: recieved pkt. Peer %s, Id: %d, Seq: %d, Recv: %s\n", peer.String(), rm.Body.(*icmp.Echo).ID, rm.Body.(*icmp.Echo).Seq, pktTime.String())
			}

			// this goroutine needs to read packets from the network as fast as possible so we can get accurate timing information.
			// if the packetChan blocks, latency measurements could start to grow.  If we changed this to a non-blocking write on the
			// channel, then packets that could not be written would appear as packet loss, which is probably worse then higher latency.
			p.packetChan <- &EchoResponse{
				Peer:     peer.String(),
				Seq:      rm.Body.(*icmp.Echo).Seq,
				Id:       rm.Body.(*icmp.Echo).ID,
				Received: pktTime,
			}
		}
	}
	if p.Debug {
		log.Printf("go-pinger: ipv6 listen loop ended.")
	}
}

func (p *Pinger) processPkt() {
	defer p.processWg.Done()
	for pkt := range p.packetChan {
		key := packetKey(pkt.Peer, pkt.Id, pkt.Seq)
		p.Lock()
		req, ok := p.inFlight[key]
		if ok {
			delete(p.inFlight, key)
		}
		if ok {
			if p.Debug {
				log.Printf("go-pinger: reply packet matches request packet. %s\n", pkt.String())
			}
			req.Received = pkt.Received
			req.WaitGroup.Done()
		} else {
			if p.Debug {
				log.Printf("go-pinger: go-pinger: unexpected echo response. %s\n", pkt.String())
			}
		}
		p.Unlock()
	}
}
