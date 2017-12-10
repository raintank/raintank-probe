package msg

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var errTooSmall = errors.New("too small")
var errFmtBinWriteFailed = "binary write failed: %q"
var errFmtUnknownFormat = "unknown format %d"
var errInvalidEvent = errors.New("invalid event definition")
var errFmtInvalidSeverity = "invalid severity level %q"

type Format uint8

const (
	FormatProbeEventJson Format = iota
	FormatProbeEventMsgp
	FormatProbeEventsMsgp
)

//go:generate msgp
type ProbeEvent struct {
	Id        string            `json:"id"`
	EventType string            `json:"event_type"`
	OrgId     int64             `json:"org_id"`
	Severity  string            `json:"severity"` // enum "INFO" "WARN" "ERROR" "OK"
	Source    string            `json:"source"`
	Timestamp int64             `json:"timestamp"`
	Message   string            `json:"message"`
	Tags      map[string]string `json:"tags"`
}

type ProbeEvents []*ProbeEvent

func (e *ProbeEvent) Validate() error {
	if e.EventType == "" || e.OrgId == 0 || e.Source == "" || e.Timestamp == 0 || e.Message == "" {
		return errInvalidEvent
	}
	switch strings.ToLower(e.Severity) {
	case "info", "ok", "warn", "error", "warning", "critical":
		// nop
	default:
		return fmt.Errorf(errFmtInvalidSeverity, e.Severity)
	}
	return nil
}

type ProbeEventJson struct {
	Id        string   `json:"id"`
	EventType string   `json:"event_type"`
	OrgId     int64    `json:"org_id"`
	Severity  string   `json:"severity"`
	Source    string   `json:"source"`
	Timestamp int64    `json:"timestamp"`
	Message   string   `json:"message"`
	Tags      []string `json:"tags"`
}

// Decode message into probeEvent.
// The message Format is:
//
// Bytes: Description
// 0    : messgae format
// 1-9  : transmit timestamp 64bit Nanosecond Epoch TS (not used)
// 10-> : message body
func ProbeEventsFromMsg(msg []byte) ([]*ProbeEvent, error) {
	if len(msg) < 9 {
		return nil, errTooSmall
	}

	format := Format(msg[0])
	if format != FormatProbeEventJson && format != FormatProbeEventMsgp && format != FormatProbeEventsMsgp {
		return nil, fmt.Errorf(errFmtUnknownFormat, format)
	}

	events := ProbeEvents(make([]*ProbeEvent, 0))
	switch format {
	case FormatProbeEventJson:
		oldFormat := &ProbeEventJson{}
		err := json.Unmarshal(msg[9:], oldFormat)
		if err != nil {
			return nil, err
		}
		//convert our []string of key:valy pairs to
		// map[string]string
		tags := make(map[string]string)
		for _, t := range oldFormat.Tags {
			parts := strings.SplitN(t, ":", 2)
			tags[parts[0]] = parts[1]
		}
		events = append(events, &ProbeEvent{
			Id:        oldFormat.Id,
			EventType: oldFormat.EventType,
			OrgId:     oldFormat.OrgId,
			Severity:  oldFormat.Severity,
			Source:    oldFormat.Source,
			Timestamp: oldFormat.Timestamp,
			Message:   oldFormat.Message,
			Tags:      tags,
		})
	case FormatProbeEventMsgp:
		e := new(ProbeEvent)
		_, err := e.UnmarshalMsg(msg[9:])
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	case FormatProbeEventsMsgp:
		_, err := events.UnmarshalMsg(msg[9:])
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf(errFmtUnknownFormat, msg[0])
	}
	return events, nil
}

func CreateProbeEventMsg(event *ProbeEvent, id int64, version Format) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, uint8(version))
	if err != nil {
		return nil, fmt.Errorf(errFmtBinWriteFailed, err)
	}
	err = binary.Write(buf, binary.BigEndian, id)
	if err != nil {
		return nil, fmt.Errorf(errFmtBinWriteFailed, err)
	}
	var msg []byte
	switch version {
	case FormatProbeEventJson:
		msg, err = json.Marshal(event)
	case FormatProbeEventMsgp:
		msg, err = event.MarshalMsg(nil)
	default:
		return nil, fmt.Errorf(errFmtUnknownFormat, version)
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal metrics payload: %s", err)
	}
	_, err = buf.Write(msg)
	if err != nil {
		return nil, fmt.Errorf(errFmtBinWriteFailed, err)
	}
	return buf.Bytes(), nil
}

func CreateProbeEventsMsg(events []*ProbeEvent) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, uint8(FormatProbeEventsMsgp))
	if err != nil {
		return nil, fmt.Errorf(errFmtBinWriteFailed, err)
	}
	id := time.Now().UnixNano()
	err = binary.Write(buf, binary.BigEndian, id)
	if err != nil {
		return nil, fmt.Errorf(errFmtBinWriteFailed, err)
	}
	var msg []byte

	msg, err = ProbeEvents(events).MarshalMsg(nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal metrics payload: %s", err)
	}
	_, err = buf.Write(msg)
	if err != nil {
		return nil, fmt.Errorf(errFmtBinWriteFailed, err)
	}
	return buf.Bytes(), nil
}
