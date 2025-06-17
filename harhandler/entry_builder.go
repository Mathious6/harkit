// Package harhandler provides functionality to build HAR entries from bogdanfinn/fhttp requests
// and responses. It allows deferred HAR construction after the request is executed, ensuring
// accurate extraction of request and response data, including body content, headers, cookies,
// and timing information.
package harhandler

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/url"
	"time"

	"github.com/Mathious6/harkit/converter"
	"github.com/Mathious6/harkit/harfile"
	http "github.com/bogdanfinn/fhttp"
)

// EntryBuilder builds a HAR entry from a bogdanfinn/fhttp request and response.
// It encapsulates all relevant data including timings, headers, cookies, and body content.
type EntryBuilder struct {
	entry *harfile.Entry
}

// NewEntry initializes an empty HAR entry with default fields and timing placeholders.
// Use AddEntry to attach the request and response data once the request is completed.
func NewEntry() *EntryBuilder {
	return &EntryBuilder{
		entry: &harfile.Entry{
			StartedDateTime: time.Now(),
			Time:            -1,
			Cache:           &harfile.Cache{},
			Timings: &harfile.Timings{
				Send:    -1,
				Wait:    -1,
				Receive: -1,
			},
		},
	}
}

// AddEntry populates the HAR entry with the given HTTP request and response.
// It clones and restores the request body to prevent side effects.
// This method should be called after the HTTP request is executed to avoid blocking.
func (b *EntryBuilder) AddEntry(req *http.Request, resp *http.Response) error {
	b.entry.Timings.Receive = float64(time.Since(b.entry.StartedDateTime).Milliseconds())

	clonedReq, err := cloneRequestPreserveBody(req)
	if err != nil {
		return err
	}

	harReq, err := converter.FromHTTPRequest(clonedReq)
	if err != nil {
		return err
	}
	harResp, err := converter.FromHTTPResponse(resp)
	if err != nil {
		return err
	}

	b.entry.Request = harReq
	b.entry.Response = harResp

	b.entry.Timings.Wait = float64(time.Since(b.entry.StartedDateTime).Milliseconds()) - b.entry.Timings.Receive
	b.entry.Time = b.entry.Timings.Total()

	return nil
}

// Build finalizes and returns the constructed HAR entry.
// If resolveIP is true, a DNS resolution will be performed on the request host.
func (b *EntryBuilder) Build(resolveIP bool) *harfile.Entry {
	if resolveIP && b.entry.Request != nil {
		b.entry.ServerIPAddress = resolveServerIPAddress(b.entry.Request.URL)
	}
	return b.entry
}

// resolveServerIPAddress performs a DNS lookup on the given URL and returns the first resolved
// IP address as a string. Returns an empty string on failure. This is a blocking operation.
func resolveServerIPAddress(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	ipAddrs, err := net.DefaultResolver.LookupIPAddr(context.Background(), parsedURL.Hostname())
	if err != nil || len(ipAddrs) == 0 {
		return ""
	}
	return ipAddrs[0].IP.String()
}

// cloneRequestPreserveBody clones an HTTP request and preserves its body by buffering the content
// into memory. Both the original and the cloned request will be reset with a fresh body reader,
// allowing for safe reuse without data loss.
func cloneRequestPreserveBody(req *http.Request) (*http.Request, error) {
	if req.Body == nil {
		return req.Clone(req.Context()), nil
	}

	buf, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(buf))

	clonedReq := req.Clone(req.Context())
	clonedReq.Body = io.NopCloser(bytes.NewReader(buf))

	return clonedReq, nil
}
