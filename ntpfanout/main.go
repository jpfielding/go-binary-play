package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type args struct {
	Hosts string
	URL   string
}

// Parse ...
func (a *args) Parse() {
	flag.StringVar(
		&a.Hosts,
		"hosts",
		"time.nist.gov:123,time.windows.com:123,time.google.com:123,time.apple.com:123,time.facebook.com:123",
		"the list of hosts",
	)
	flag.StringVar(
		&a.URL,
		"url",
		"https://gist.githubusercontent.com/mutin-sa/eea1c396b1e610a2da1e5550d94b0453/raw/f25741933f7729d63fcd23e6d9fc2ff1cddc170a/Public_Time_Servers.md",
		"the url to containg a line by line list of time servers",
	)
	flag.Parse()
}

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
	// container for params
	args := &args{}
	args.Parse()

	// register sig int/term
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)

	// register the primary context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// hosts input
	hosts := make(chan string, 3)
	go func() {
		defer close(hosts)
		for _, h := range strings.Split(args.Hosts, ",") {
			hosts <- h
		}
		if args.URL != "" {
			client := &http.Client{}
			rsp, err := client.Get(args.URL)
			if err != nil {
				cancel()   // kill the context
				panic(err) // never in production
			}
			defer rsp.Body.Close()
			scanner := bufio.NewScanner(rsp.Body)
			// optionally, resize scanner's capacity for lines over 64K, see next example
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				switch {
				case strings.HasSuffix(line, ":"), strings.HasPrefix(line, "#"), line == "":
					continue
				}
				hosts <- line
			}
		}
	}()

	// a chan for responses
	responses := make(chan response)
	var wg sync.WaitGroup
	for h := range hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
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
	if !strings.Contains(host, ":") {
		host = host + ":123"
	}
	conn, err := net.Dial("udp", host)
	if err != nil {
		return response{err: fmt.Errorf("failed to connect: %w", err)}
	}
	defer conn.Close()
	duration := time.Now().Add(15 * time.Second)
	if err := conn.SetDeadline(duration); err != nil {
		return response{err: fmt.Errorf("failed to set deadline: %w", err)}
	}

	// configure request settings by specifying the first byte as
	// 00 011 011 (or 0x1B)
	// |  |   +-- client mode (3)
	// |  + ----- version (3)
	// + -------- leap year indicator, 0 no warning
	req := &ntpv4{Settings: 0x1B}

	if err := binary.Write(conn, binary.BigEndian, req); err != nil {
		return response{err: fmt.Errorf("failed to send request: %w, %w", err, req)}
	}

	rsp := ntpv4{}
	if err := binary.Read(conn, binary.BigEndian, &rsp); err != nil {
		return response{err: fmt.Errorf("failed to read server response: %w %v", err, req)}
	}

	return response{host: host, ntp: rsp, err: nil}
}
