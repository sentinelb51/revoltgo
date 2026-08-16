package revoltgo

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/quic-go/quic-go/http3"
)

// probeTimeout caps how long the background HTTP/3 probe may take.
const probeTimeout = 10 * time.Second

/*
h3Transport sends over HTTP/3 where it works and TCP everywhere else.

A host starts on TCP. If it advertises h3 in Alt-Svc, a background probe checks
that UDP actually gets through; only then do real requests move to QUIC. That
ordering matters: a blocked UDP path costs a throwaway probe rather than
someone's message, so no request is ever replayed to discover the protocol.

Decisions are per-host and sticky. The library talks to both the API and the
CDN, which need not support h3 alike.
*/
type h3Transport struct {
	tcp  *http.Transport
	quic *http3.Transport

	// hosts where HTTP/3 can be used if `bool` true. Missing entry means host wasn't probed.
	hosts sync.Map
}

func newH3Transport() *h3Transport {
	// Copy our own transport to avoid the process-global http.DefaultTransport; no connection-pool sharing.
	tcp := http.DefaultTransport.(*http.Transport).Clone()

	// HTTP/2 or nothing; a server that won't negotiate h2 fails the request
	// rather than silently downgrading. Protocols supersedes ForceAttemptHTTP2.
	var protocols http.Protocols
	protocols.SetHTTP2(true)
	tcp.Protocols = &protocols

	return &h3Transport{
		tcp:  tcp,
		quic: &http3.Transport{},
	}
}

func (t *h3Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	host := request.URL.Host

	if usable, ok := t.hosts.Load(host); ok && usable.(bool) {
		response, err := t.quic.RoundTrip(request)
		if err == nil {
			return response, nil
		}

		// QUIC broke mid-session: revert this host to TCP for good, then retry
		// here only if the body can be rewound.
		t.hosts.Store(host, false)
		log.Printf("HTTP/3 to %s failed (%v), reverting to TCP", host, err)

		if request.Body != nil {
			if request.GetBody == nil {
				return nil, err
			}

			if request.Body, err = request.GetBody(); err != nil {
				return nil, err
			}
		}
	}

	response, err := t.tcp.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	if request.URL.Scheme == "https" && strings.Contains(response.Header.Get("Alt-Svc"), "h3=") {
		t.probe(host)
	}

	return response, nil
}

// probe confirms in the background that h3 reaches a host, and leaves the
// connection warmed in the pool for the requests that follow.
func (t *h3Transport) probe(host string) {
	// Claim the probe; false doubles as "in flight" and "failed".
	if _, probed := t.hosts.LoadOrStore(host, false); probed {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
		defer cancel()

		request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+host+"/", http.NoBody)
		if err != nil {
			return
		}

		// Any response proves the path works; the status code is irrelevant.
		response, err := t.quic.RoundTrip(request)
		if err != nil {
			log.Printf("HTTP/3 unavailable for %s (%v), staying on TCP", host, err)
			return
		}

		// Nothing to read; closing resets the stream, the connection stays warm.
		_ = response.Body.Close()
		t.hosts.Store(host, true)
		log.Printf("%s compatible with QUIC; upgrading to HTTP/3", host)
	}()
}

// Close releases both connection pools. The QUIC one holds a UDP socket and a
// reader goroutine, so it does not go away on its own.
func (t *h3Transport) Close() error {
	t.tcp.CloseIdleConnections()
	return t.quic.Close()
}
