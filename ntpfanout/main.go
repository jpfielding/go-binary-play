package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
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
	hosts := []string{
		"time.nist.gov:123",
		"time.windows.com:123",
		"time.google.com:123",
		"time.apple.com:123",
		"time.facebook.com:123",
	}

	// register sig int/term
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)

	// register the primary context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// a chan for responses
	responses := make(chan response)
	var wg sync.WaitGroup
	for _, h := range hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			defer fmt.Printf("%s is complete\n", host)
			select {
			case <-ctx.Done():
				return // canceled
			case <-sigc:
				return // sig
			case <-time.NewTimer(1 * time.Second).C:
				return // t/o
			case responses <- timeFrom(host):
				return // got a response
			}
		}(h)
	}
	// listen in another gorourtine for being complete
	go func() {
		// ensure that all goroutines were startted
		wg.Wait()
		// no more writing to responses
		close(responses)
	}()
	// read out our responses
	for rsp := range responses {
		const ntpEpochOffset = 2208988800
		secs := float64(rsp.ntp.TxTimeSec) - ntpEpochOffset
		nanos := (int64(rsp.ntp.TxTimeFrac) * 1e9) >> 32
		fmt.Printf("%s %v %v\n", rsp.host, time.Unix(int64(secs), nanos), rsp.err)
	}
}

type response struct {
	host string
	ntp  ntpv4
	err  error
}

func timeFrom(host string) response {
	conn, err := net.Dial("udp", host)
	if err != nil {
		return response{err: errors.Wrap(err, "failed to connect:")}
	}
	defer conn.Close()
	duration := time.Now().Add(15 * time.Second)
	if err := conn.SetDeadline(duration); err != nil {
		return response{err: errors.Wrap(err, "failed to set deadline: ")}
	}

	// configure request settings by specifying the first byte as
	// 00 011 011 (or 0x1B)
	// |  |   +-- client mode (3)
	// |  + ----- version (3)
	// + -------- leap year indicator, 0 no warning
	req := &ntpv4{Settings: 0x1B}

	if err := binary.Write(conn, binary.BigEndian, req); err != nil {
		return response{err: errors.Wrap(err, "failed to send request: %v")}
	}

	rsp := ntpv4{}
	if err := binary.Read(conn, binary.BigEndian, &rsp); err != nil {
		return response{err: errors.Wrap(err, "failed to read server response: %v")}
	}

	return response{host: host, ntp: rsp, err: nil}
}
