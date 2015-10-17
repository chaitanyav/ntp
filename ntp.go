package ntp

//Reference https://www.eecis.udel.edu/~mills/database/rfc/rfc1305/rfc1305c.pdf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
  "time"
)

var leapIndicator map[byte]string
var mode map[byte]string
var version byte

type NTP interface {
	decodeStratum() string
	decodeLeapIndicator() string
	decodeVersion() byte
	decodeMode() string
  decodeReferenceIdentifier() string
}

type DataPacket struct {
	Byte1               byte
	Stratum             byte
	Poll                int8
	Precision           int8
	RootDelay           int32
	RootDispersion      int32
	ReferenceIdentifier uint32
	ReferenceTimeStamp  uint64
	OriginateTimeStamp  uint64
	ReceiveTimeStamp    uint64
	TransmitTimeStamp   uint64
}

func init() {
	leapIndicator = make(map[byte]string)
	leapIndicator[0] = "no warning"
	leapIndicator[1] = "last minute has 61 seconds"
	leapIndicator[2] = "last minute has 59 seconds"
	leapIndicator[3] = "alarm condition(clock not synchronized)"

	mode = make(map[byte]string)
	mode[0] = "reserved"
	mode[1] = "symmetric active"
	mode[2] = "symmetric passive"
	mode[3] = "client"
	mode[4] = "server"
	mode[5] = "broadcast"
	mode[6] = "reserved for ntp control message"
	mode[7] = "reserved for private use"
}

func (packet *DataPacket) decodeStratum() string {
  stratum := ""
  if packet.Stratum == 0 {
    stratum = "unspecified"
  } else if packet.Stratum == 1 {
    stratum = "primary reference (e.g radio clock)"
  } else if packet.Stratum >= 2 && packet.Stratum <= 255 {
    stratum = "secondary reference (via NTP)"
  }
  return stratum
}

func (packet *DataPacket) decodeLeapIndicator() string {
	b := (packet.Byte1 >> 6) & 3
	return leapIndicator[b]
}

func (packet *DataPacket) decodeVersion() byte {
	return (packet.Byte1 >> 3) & 7
}

func (packet *DataPacket) decodeMode() string {
	b := packet.Byte1 & 7
  return mode[b]
}

func (packet *DataPacket) decodeReferenceIdentifier() string {
    return ""
}

func query(packet DataPacket) (*DataPacket, error) {
	conn, err := net.Dial("udp", "0.pool.ntp.org:123")
	if err != nil {
		log.Printf("error on connecting to NTP Server: %v\n", err)
		return nil, err
	}
	tmpBuf := new(bytes.Buffer)
	err = binary.Write(tmpBuf, binary.BigEndian, packet)
	if err != nil {
		log.Printf("error on converting the packet to bytes: %v\n", err)
		return nil, err
	}

	n, err := conn.Write(tmpBuf.Bytes())
	if err != nil {
		log.Printf("error on writing to UDP socket: %v\n", err)
		return nil, err
	}
	log.Printf("Sent %d bytes on UDP socket\n", n)

	data := make([]byte, 48)
	n, err = conn.Read(data)
	if err != nil {
		log.Printf("error on reading from UDP socket: %v\n", err)
		return nil, err
	}

	log.Printf("Received %d bytes on UDP socket\n", n)
	outBuf := bytes.NewReader(data)
	resPacket := DataPacket{}
	err = binary.Read(outBuf, binary.BigEndian, &resPacket)
	if err != nil {
		log.Printf("error converting the response to packet: %v\n", err)
		return nil, err
	}
	return &resPacket, nil
}
