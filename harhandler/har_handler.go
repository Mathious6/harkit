package harhandler

import (
	"context"
	"fmt"
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
	globalHarStorage      = make(map[string]*HARHandler) // globalHarStorage is storing all the HARHandlers for all the flows
	globalHarStorageMutex = sync.Mutex{}                 // globalHarStorageMutex is used to synchronize access to the globalHarStorage map
)

// HARHandler is the main struct that stores the HAR data for a flow
type HARHandler struct {
	har              *harfile.HAR // har is the HAR data for the flow
	resolveIPAddress bool         // resolveIPAddress is a flag to resolve the IP address of the server
}

// CreateHandler creates a new HARHandler for a flow with the given flowID
func CreateHandler(flowID string, opts ...HandlerOption) (*HARHandler, error) {
	globalHarStorageMutex.Lock()
	defer globalHarStorageMutex.Unlock()
	if _, exists := globalHarStorage[flowID]; exists {
		return nil, fmt.Errorf("handler %q already exists", flowID)
	}
	handler := newHARHandler(flowID, opts...)
	globalHarStorage[flowID] = handler
	return handler, nil
}

// GetHandler gets the HARHandler for a flow with the given flowID
func GetHandler(flowID string) (*HARHandler, error) {
	globalHarStorageMutex.Lock()
	defer globalHarStorageMutex.Unlock()
	handler, exists := globalHarStorage[flowID]
	if !exists {
		return nil, fmt.Errorf("handler %q not found", flowID)
	}
	return handler, nil
}

// GetOrCreateHandler gets the HARHandler for a flow with the given flowID, if it doesn't exist, it creates a new one
func GetOrCreateHandler(flowID string, opts ...HandlerOption) *HARHandler {
	if handler, err := GetHandler(flowID); err == nil {
		for _, opt := range opts {
			opt(handler)
		}
		return handler
	}
	handler, _ := CreateHandler(flowID, opts...)
	return handler
}

// newHARHandler creates a new HARHandler for a flow with the given flowID and applies the given options
func newHARHandler(flowID string, opts ...HandlerOption) *HARHandler {
	h := &HARHandler{
		har: &harfile.HAR{
			Log: &harfile.Log{
				Version: harfile.HARVersion,
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

// AddEntry adds a new entry to the HARHandler for a flow with the given flowID, sentAt, request, and response
func AddEntry(flowId string, proxy string, sentAt time.Time, req *http.Request, resp *http.Response) error {
	return GetOrCreateHandler(flowId).AddEntry(proxy, sentAt, req, resp)
}

// Export exports the HAR data for a flow with the given flowID and filename
func Export(flowId, filename string) error {
	handler := GetOrCreateHandler(flowId)
	delete(globalHarStorage, flowId)
	return handler.har.Save(filename)
}

// AddEntry adds a new entry to the HARHandler for a flow with the given sentAt, request, and response
func (h *HARHandler) AddEntry(proxy string, sentAt time.Time, req *http.Request, resp *http.Response) error {
	timingsReceive := float64(time.Since(sentAt).Milliseconds())

	harReq, err := converter.FromHTTPRequest(req)
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
		StartedDateTime: harfile.HARTime(sentAt),
		Time:            timings.Total(),
		Request:         harReq,
		Response:        harResp,
		Cache:           nil,
		Timings:         timings,
		ClientProxy:     proxy,
		ServerIPAddress: resolveServerIPAddress(h.resolveIPAddress, harReq.URL),
	})

	return nil
}

// resolveServerIPAddress performs a DNS lookup on the given URL and returns the first resolved
// IP address as a string. Returns an empty string on failure. This is a blocking operation.
func resolveServerIPAddress(resolve bool, rawURL string) string {
	if !resolve {
		return "0.0.0.0"
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
