package harhandler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/Mathious6/harkit"
	"github.com/Mathious6/harkit/converter"
	"github.com/Mathious6/harkit/harfile"
	http "github.com/bogdanfinn/fhttp"
)

var (
	globalHarStorage      = make(map[string]*HARHandler)
	globalHarStorageMutex = sync.Mutex{}
)

type HARHandler struct {
	har *harfile.HAR

	resolveIPAddress bool
}

func CreateHandler(flowID string, opts ...HandlerOption) (*HARHandler, error) {
	globalHarStorageMutex.Lock()
	defer globalHarStorageMutex.Unlock()
	if _, ok := globalHarStorage[flowID]; ok {
		return nil, fmt.Errorf("handler %q already exists", flowID)
	}
	handler := newHARHandler(flowID, opts...)
	globalHarStorage[flowID] = handler
	return handler, nil
}

func GetHandler(flowID string) (*HARHandler, error) {
	globalHarStorageMutex.Lock()
	defer globalHarStorageMutex.Unlock()
	handler, exists := globalHarStorage[flowID]
	if !exists {
		return nil, fmt.Errorf("handler %q not found", flowID)
	}
	return handler, nil
}

func GetOrCreateHandler(flowID string, opts ...HandlerOption) *HARHandler {
	if h, err := GetHandler(flowID); err == nil {
		for _, opt := range opts {
			opt(h)
		}
		return h
	}
	h, _ := CreateHandler(flowID, opts...)
	return h
}

func newHARHandler(flowID string, opts ...HandlerOption) *HARHandler {
	h := &HARHandler{
		har: &harfile.HAR{
			Log: &harfile.Log{
				Version: "1.2",
				Creator: &harfile.Creator{
					Name:    flowID,
					Version: fmt.Sprintf("harkit-%s", harkit.Version),
				},
				Entries: []*harfile.Entry{},
			},
		},
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func AddEntry(flowId string, sentAt time.Time, req *http.Request, resp *http.Response) error {
	handler := GetOrCreateHandler(flowId)
	return handler.AddEntry(sentAt, req, resp)
}

func Export(flowId string, filename string) error {
	handler := GetOrCreateHandler(flowId)
	delete(globalHarStorage, flowId)
	return handler.har.Save(filename)
}

func (h *HARHandler) AddEntry(sentAt time.Time, req *http.Request, resp *http.Response) error {
	timingsReceive := float64(time.Since(sentAt).Milliseconds())

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

	timings := &harfile.Timings{
		Send:    -1,
		Wait:    float64(time.Since(sentAt).Milliseconds()) - timingsReceive,
		Receive: timingsReceive,
	}

	h.har.Log.Entries = append(h.har.Log.Entries, &harfile.Entry{
		StartedDateTime: sentAt,
		Time:            timings.Total(),
		Request:         harReq,
		Response:        harResp,
		Cache:           &harfile.Cache{},
		Timings:         timings,
		ServerIPAddress: resolveServerIPAddress(h.resolveIPAddress, harReq.URL),
	})

	return nil
}

// resolveServerIPAddress performs a DNS lookup on the given URL and returns the first resolved
// IP address as a string. Returns an empty string on failure. This is a blocking operation.
func resolveServerIPAddress(resolve bool, rawURL string) string {
	if !resolve {
		return ""
	}

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
