package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

type ntpv4 struct {
	Settings       uint8  // leap yr indicator, ver number, and mode
	Stratum        uint8  // stratum of local clock
	Poll           int8   // poll exponent
	Precision      int8   // precision exponent
	RootDelay      uint32 // root delay
	RootDispersion uint32 // root dispersion
	ReferenceID    uint32 // reference id
	RefTimeSec     uint32 // reference timestamp sec
	RefTimeFrac    uint32 // reference timestamp fractional
	OrigTimeSec    uint32 // origin time secs
	OrigTimeFrac   uint32 // origin time fractional
	RxTimeSec      uint32 // receive time secs
	RxTimeFrac     uint32 // receive time frac
	TxTimeSec      uint32 // transmit time secs
	TxTimeFrac     uint32 // transmit time frac}
}

// https://datatracker.ietf.org/doc/html/rfc5905
func main() {
	host := "time.nist.gov:123"
	conn, err := net.Dial("udp", host)
	if err != nil {
		log.Fatal("failed to connect:", err)
	}
	defer conn.Close()
	duration := time.Now().Add(15 * time.Second)
	if err := conn.SetDeadline(duration); err != nil {
		log.Fatal("failed to set deadline: ", err)
	}

	// configure request settings by specifying the first byte as
	// 00 011 011 (or 0x1B)
	// |  |   +-- client mode (3)
	// |  + ----- version (3)
	// + -------- leap year indicator, 0 no warning
	req := &ntpv4{Settings: 0x1B}

	if err := binary.Write(conn, binary.BigEndian, req); err != nil {
		log.Fatalf("failed to send request: %v", err)
	}

	rsp := &ntpv4{}
	if err := binary.Read(conn, binary.BigEndian, rsp); err != nil {
		log.Fatalf("failed to read server response: %v", err)
	}

	fmt.Printf("%v\n", rsp)

	const ntpEpochOffset = 2208988800
	secs := float64(rsp.TxTimeSec) - ntpEpochOffset
	nanos := (int64(rsp.TxTimeFrac) * 1e9) >> 32
	fmt.Printf("%v\n", time.Unix(int64(secs), nanos))
}
