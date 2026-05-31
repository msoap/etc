package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptrace"
	"os"
	"sort"
	"time"

	"github.com/msoap/tcg"
)

/*

http-trace is a simple utility to trace HTTP requests and responses.

usage:

	go run http-trace.go <url>

	# or after installing binary:
	http-trace <url>

install:

	go install github.com/msoap/etc/http-trace@latest

example:

  $ go run http-trace.go 'https://pkg.go.dev/'
  Tracing HTTP requests to https://pkg.go.dev/
  Response body length: 33470
  start       : 10:48:21.574,           0s ▖
  dnsStart    : 10:48:21.575,    496.084µs ▗
  dnsDone     : 10:48:21.630,  55.200416ms  ▄▄▄▄▄▄▄▄▖
  connectStart: 10:48:21.630,    105.417µs          ▗
  connectDone : 10:48:21.655,  25.305125ms           ▄▄▄▄
  tlsStart    : 10:48:21.655,     42.292µs               ▖
  tlsDone     : 10:48:21.699,  43.511083ms               ▗▄▄▄▄▄▄
  gotConn     : 10:48:21.699,     378.25µs                      ▖
  wroteHeaders: 10:48:21.699,      179.5µs                      ▗
  firstByte   : 10:48:21.863, 163.526458ms                       ▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▖
  fullDone    : 10:48:21.889,  25.874625ms                                                ▗▄▄
  Total time  : 314.61925ms

*/

type timing struct {
	name string
	time time.Time
}

type timings []timing

const (
	startKey        = "start"
	dnsStartKey     = "dnsStart"
	dnsDoneKey      = "dnsDone"
	connectStartKey = "connectStart"
	connectDoneKey  = "connectDone"
	wroteHeadersKey = "wroteHeaders"
	tlsStartKey     = "tlsStart"
	tlsDoneKey      = "tlsDone"
	gotConnKey      = "gotConn"
	firstByteKey    = "firstByte"
	fullDoneKey     = "fullDone"
)

func (t *timings) addStart() {
	*t = append(*t, timing{name: startKey, time: time.Now()})
}

func (t *timings) addDNSStart() {
	*t = append(*t, timing{name: dnsStartKey, time: time.Now()})
}

func (t *timings) addDNSDone() {
	*t = append(*t, timing{name: dnsDoneKey, time: time.Now()})
}

func (t *timings) addConnectStart() {
	*t = append(*t, timing{name: connectStartKey, time: time.Now()})
}

func (t *timings) addConnectDone() {
	*t = append(*t, timing{name: connectDoneKey, time: time.Now()})
}

func (t *timings) addWroteHeaders() {
	*t = append(*t, timing{name: wroteHeadersKey, time: time.Now()})
}

func (t *timings) addTLSStart() {
	*t = append(*t, timing{name: tlsStartKey, time: time.Now()})
}

func (t *timings) addTLSDone() {
	*t = append(*t, timing{name: tlsDoneKey, time: time.Now()})
}

func (t *timings) addGotConn() {
	*t = append(*t, timing{name: gotConnKey, time: time.Now()})
}

func (t *timings) addFirstByte() {
	*t = append(*t, timing{name: firstByteKey, time: time.Now()})
}

func (t *timings) addEndResponse() {
	*t = append(*t, timing{name: fullDoneKey, time: time.Now()})
}

func (t *timings) print() {
	const (
		format     = "15:04:05.000"
		chartWidth = 50
	)
	// sort timings by time
	sort.Slice(*t, func(i, j int) bool {
		return (*t)[i].time.Before((*t)[j].time)
	})

	fullSpent := (*t)[len(*t)-1].time.Sub((*t)[0].time)
	spentBeforePixels := 0
	chartLine := ""
	for i, timing := range *t {
		prevI := max(i-1, 0)
		spentToPrev := timing.time.Sub((*t)[prevI].time)
		spentBeforePixels, chartLine = renderChartLine(spentBeforePixels, spentToPrev, fullSpent, chartWidth)

		fmt.Printf("%-12s: %s, %12s %s\n",
			timing.name,
			timing.time.Format(format),
			spentToPrev,
			chartLine,
		)
	}
	fmt.Printf("%-12s: %s\n", "Total time", fullSpent)
}

func renderChartLine(spentBeforePixels int, spentToPrev, fullSpent time.Duration, width int) (int, string) {
	// |1|1|1|2|2| | | | | |
	//  1 - spent before (empty pixels)
	//  2 - spent to prev
	//  + others: all

	bufWidth := width * 2
	if bufWidth <= 0 {
		return spentBeforePixels, ""
	}

	tbuf := tcg.NewBuffer(bufWidth, 2)

	if fullSpent <= 0 {
		tbuf.HLine(0, 0, 1, 1)
		rendered := tbuf.RenderAsStrings(tcg.Mode2x2)
		if len(rendered) < 1 {
			log.Printf("unexpected empty rendered chart")
			return spentBeforePixels, ""
		}
		return min(spentBeforePixels+1, bufWidth), rendered[0]
	}

	durPerPixel := float64(fullSpent) / float64(bufWidth)
	spentInPixels := int(float64(spentToPrev) / durPerPixel)

	if spentInPixels == 0 {
		spentInPixels = 1
	}
	if spentInPixels > bufWidth {
		spentInPixels = bufWidth
	}
	if spentBeforePixels >= bufWidth {
		spentBeforePixels = bufWidth - 1
	}

	tbuf.HLine(spentBeforePixels, 1, spentInPixels, 1)

	rendered := tbuf.RenderAsStrings(tcg.Mode2x2)
	if len(rendered) < 1 {
		log.Printf("unexpected empty rendered chart")
		return 0, ""
	}

	return min(spentBeforePixels+spentInPixels, bufWidth), rendered[0]
}

func main() {
	if len(os.Args) < 2 {
		println("Usage: go run http-trace.go <url>")
		os.Exit(1)
	}
	link := os.Args[1]
	fmt.Printf("Tracing HTTP requests to %s\n", link)

	tm := &timings{}
	tm.addStart()

	req, _ := http.NewRequest("GET", link, nil)
	trace := &httptrace.ClientTrace{
		GotConn: func(_ httptrace.GotConnInfo) {
			tm.addGotConn()
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			tm.addDNSDone()
		},
		DNSStart: func(_ httptrace.DNSStartInfo) {
			tm.addDNSStart()
		},
		ConnectStart: func(_, _ string) {
			tm.addConnectStart()
		},
		GotFirstResponseByte: func() {
			tm.addFirstByte()
		},
		ConnectDone: func(_, _ string, _ error) {
			tm.addConnectDone()
		},
		WroteHeaders: func() {
			tm.addWroteHeaders()
		},
		TLSHandshakeStart: func() {
			tm.addTLSStart()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			tm.addTLSDone()
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	if resp.Body != nil {
		// read all response body
		if body, err := io.ReadAll(resp.Body); err != nil {
			log.Printf("Failed to read response body: %v", err)
		} else {
			fmt.Printf("Response body length: %d\n", len(body))
		}

		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}

	tm.addEndResponse()
	tm.print()
}
