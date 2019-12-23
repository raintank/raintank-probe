package pinger

import (
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/net/icmp"
)

// An EchoRequest is a single ICMP echo request packet that will be sent.
// When the echoResponse is recieved, the "Received" field will be set and
// the Done() will be called on the WaitGroup.
//
type EchoRequest struct {
	icmp.Message

	Destination net.IP
	Sent        time.Time
	Received    time.Time
	ID          string
	WaitGroup   *sync.WaitGroup
}

// create a new EchoRequest instance.
func NewEchoRequest(msg icmp.Message, dest net.IP, wg *sync.WaitGroup) *EchoRequest {
	return &EchoRequest{
		Message:     msg,
		Destination: dest,
		WaitGroup:   wg,
		ID:          packetKey(dest.String(), msg.Body.(*icmp.Echo).ID, msg.Body.(*icmp.Echo).Seq),
	}

}

func packetKey(addr string, id, seq int) string {
	return fmt.Sprintf("%s-%d-%d", addr, id, seq)
}

func (p *Pinger) Send(req *EchoRequest) error {
	b, err := req.Marshal(nil)
	if err != nil {
		return err
	}
	req.Sent = time.Now()
	if req.Destination.To4() == nil {
		_, err = p.v6Conn.WriteTo(b, &net.IPAddr{IP: req.Destination})
		if err != nil {
			return err
		}
	} else {
		_, err = p.v4Conn.WriteTo(b, &net.IPAddr{IP: req.Destination})
		if err != nil {
			return err
		}
	}
	return nil
}
