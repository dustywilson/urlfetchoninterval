package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/gentlemanautomaton/signaler"
)

type options struct {
	URL      *url.URL          `kong:"env='FETCH_URL',required,arg"`
	Interval time.Duration     `kong:"env='FETCH_INTERVAL',default='1m',short='i'"`
	Timeout  time.Duration     `kong:"env='FETCH_TIMEOUT',default='5s',short='t'"`
	Headers  map[string]string `kong:"env='FETCH_HEADERS'"`
	Proxy    *url.URL          `kong:"env='PROXY_URL',short='p'"`
	Verbose  bool              `kong:"env='VERBOSE',short='v'"`
}

func (opts *options) Validate() error {
	if u := opts.URL; u != nil {
		if !u.IsAbs() {
			return fmt.Errorf("the \"%s\" URL is not absolute", u)
		}
		if u.Hostname() == "" {
			return fmt.Errorf("failed to identify a hostname within \"%s\"", u)
		}
	}
	if u := opts.Proxy; u != nil {
		if !u.IsAbs() {
			return fmt.Errorf("the \"%s\" proxy URL is not absolute", u)
		}
		if u.Hostname() == "" {
			return fmt.Errorf("failed to identify a hostname within \"%s\"", u)
		}
	}
	if opts.Interval < time.Second {
		return errors.New("the provided fetch interval is less than one second")
	}
	return nil
}

func (opts *options) Summary() (lines []string) {
	lines = append(lines, fmt.Sprintf("URL: %s", opts.URL))
	lines = append(lines, fmt.Sprintf("INTERVAL: %s", opts.Interval))
	lines = append(lines, fmt.Sprintf("TIMEOUT: %s", opts.Timeout))
	if opts.Proxy != nil {
		lines = append(lines, fmt.Sprintf("PROXY: %s", opts.Proxy))
	}
	if opts.Verbose {
		lines = append(lines, "VERBOSE")
	}
	for key, value := range opts.Headers {
		lines = append(lines, fmt.Sprintf("%s: %s", key, value))
	}
	return
}

func main() {
	var opts options

	// Parse options
	kong.Parse(&opts,
		kong.Description("Fetches a URL on an interval."),
		kong.UsageOnError())

	// Don't let requests last longer than the desired interval
	if opts.Timeout > opts.Interval {
		opts.Timeout = opts.Interval
	}

	// Capture shutdown signals
	shutdown := signaler.New().Capture(os.Interrupt, syscall.SIGTERM)
	defer shutdown.Trigger()
	ctx := shutdown.Context()

	// Prepare an HTTP client
	client := newClient(opts.Timeout, opts.Proxy)

	// Announce startup
	fmt.Printf("The process has started with this configuration:\n  %s\n", strings.Join(opts.Summary(), "\n  "))

	// Fetch on an interval until we receive a shutdown signal
	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("The process has stopped.")
			return
		case <-ticker.C:
			fetch(ctx, client, opts.URL, opts.Headers, opts.Verbose)
		}
	}
}

func newClient(timeout time.Duration, proxy *url.URL) *http.Client {
	// Prepare a transport that has the requested timeout
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: timeout,
		}).Dial,
		TLSHandshakeTimeout: timeout,
	}

	// If a proxy has been specified, configure the transport to use it
	if proxy != nil {
		transport.Proxy = http.ProxyURL(proxy)
	}

	// Return the client with the prepared transport and requested timeout
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

func fetch(ctx context.Context, c *http.Client, u *url.URL, headers map[string]string, verbose bool) {
	// Prepare an HTTP request with the given URL and headers
	req, err := http.NewRequest("GET", u.String(), nil)
	for key, value := range headers {
		req.Header[key] = []string{value}
	}

	// Include the given context for cancellation (this facilitates CTRL+C)
	req = req.WithContext(ctx)

	// Execute the HTTP request
	resp, err := c.Do(req)
	if err != nil {
		fmt.Printf("FETCH %s %v\n", u, err)
		return
	}
	defer resp.Body.Close() // This will drain resp.Body if necessary

	// If verbose output has been requested, dump the body to stdout
	if verbose {
		fmt.Printf("FETCH %s %v\n------------\n", u, resp.Status)
		io.Copy(os.Stdout, resp.Body)
		fmt.Printf("------------\n")
	} else {
		fmt.Printf("FETCH %s %v\n", u, resp.Status)
	}
}
