// Package harfile provides types for working with HAR (HTTP Archive) 1.2 files.
//
// See: http://www.softwareishard.com/blog/har-12-spec/
package harfile

import (
	"encoding/json"
	"os"
	"time"
)

const (
	HARVersion    = "1.2"                           // HARVersion is the version of the HAR format
	HARTimeLayout = "2006-01-02T15:04:05.000-07:00" // REF: "Mon Jan 2 15:04:05 MST 2006"
)

// HARTime is a wrapper around time.Time that formats the time in the HAR format.
type HARTime time.Time

func (ht HARTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(ht).Format(HARTimeLayout) + `"`), nil
}

// HAR parent container for log.
type HAR struct {
	Log *Log `json:"log"` // Log represents the root of exported data.
}

// Log represents the root of exported data.
type Log struct {
	Version string   `json:"version"`           // Version number of the format. If empty, string "1.1" is assumed by default.
	Creator *Creator `json:"creator"`           // Name and version info of the log creator application.
	Browser *Browser `json:"browser,omitempty"` // Name and version info of used browser.
	Pages   []*Page  `json:"pages,omitempty"`   // List of all exported (tracked) pages. Leave out this field if the application does not support grouping by pages.
	Entries []*Entry `json:"entries"`           // List of all exported (tracked) requests.
	Comment string   `json:"comment,omitempty"` // A comment provided by the user or the application.
}

// Creator creator and browser objects share the same structure.
type Creator struct {
	Name    string `json:"name"`              // Name of the application/browser used to export the log.
	Version string `json:"version"`           // Version of the application/browser used to export the log.
	Comment string `json:"comment,omitempty"` // A comment provided by the user or the application.
}

// Browser browser and creator objects share the same structure.
type Browser struct {
	Name    string `json:"name"`              // Name of the application/browser used to export the log.
	Version string `json:"version"`           // Version of the application/browser used to export the log.
	Comment string `json:"comment,omitempty"` // A comment provided by the user or the application.
}

// Pages represents list of exported pages.
type Page struct {
	StartedDateTime HARTime      `json:"startedDateTime"`   // Date and time stamp for the beginning of the page load (ISO 8601 - YYYY-MM-DDThh:mm:ss.sTZD, e.g. 2009-07-24T19:20:30.45+01:00).
	ID              string       `json:"id"`                // Unique identifier of a page within the [log]. Entries use it to refer the parent page.
	Title           string       `json:"title"`             // Page title.
	PageTimings     *PageTimings `json:"pageTimings"`       // Detailed timing info about page load.
	Comment         string       `json:"comment,omitempty"` // A comment provided by the user or the application.
}

// PageTimings describes timings for various events (states) fired during the page load. All times
// are specified in milliseconds. If a time info is not available appropriate field is set to -1.
type PageTimings struct {
	OnContentLoad float64 `json:"onContentLoad,omitempty,omitzero"` // Content of the page loaded. Number of milliseconds since page load started (page.startedDateTime). Use -1 if the timing does not apply to the current request.
	OnLoad        float64 `json:"onLoad,omitempty,omitzero"`        // Page is loaded (onLoad event fired). Number of milliseconds since page load started (page.startedDateTime). Use -1 if the timing does not apply to the current request.
	Comment       string  `json:"comment,omitempty"`                // A comment provided by the user or the application.
}

// Entry represents an array with all exported HTTP requests. Sorting entries by startedDateTime
// (starting from the oldest) is preferred way how to export data since it can make importing
// faster. However the reader application should always make sure the array is sorted (if required
// for the import).
type Entry struct {
	Pageref         string    `json:"pageref,omitempty"`         // Reference to the parent page. Leave out this field if the application does not support grouping by pages.
	StartedDateTime HARTime   `json:"startedDateTime"`           // Date and time stamp of the request start (ISO 8601 - YYYY-MM-DDThh:mm:ss.sTZD).
	Time            float64   `json:"time"`                      // Total elapsed time of the request in milliseconds. This is the sum of all timings available in the timings object (i.e. not including -1 values) .
	Request         *Request  `json:"request"`                   // Detailed info about the request.
	Response        *Response `json:"response"`                  // Detailed info about the response.
	Cache           *Cache    `json:"cache"`                     // Info about cache usage.
	Timings         *Timings  `json:"timings"`                   // Detailed timing info about request/response round trip.
	ClientProxy     string    `json:"clientProxy,omitempty"`     // NEW: Proxy used by the client to connect to the server.
	ServerIPAddress string    `json:"serverIPAddress,omitempty"` // IP address of the server that was connected (result of DNS resolution).
	Connection      string    `json:"connection,omitempty"`      // Unique ID of the parent TCP/IP connection, can be the client or server port number. Note that a port number doesn't have to be unique identifier in cases where the port is shared for more connections. If the port isn't available for the application, any other unique connection ID can be used instead (e.g. connection index). Leave out this field if the application doesn't support this info.
	Comment         string    `json:"comment,omitempty"`         // A comment provided by the user or the application.
}

// Request contains detailed info about performed request.
type Request struct {
	Method      string    `json:"method"`             // Request method (GET, POST, ...).
	URL         string    `json:"url"`                // Absolute URL of the request (fragments are not included).
	HTTPVersion string    `json:"httpVersion"`        // Request HTTP Version.
	Cookies     []*Cookie `json:"cookies"`            // List of cookie objects.
	Headers     []*NVPair `json:"headers"`            // List of header objects.
	QueryString []*NVPair `json:"queryString"`        // List of query parameter objects.
	PostData    *PostData `json:"postData,omitempty"` // Posted data info.
	HeadersSize int64     `json:"headersSize"`        // Total number of bytes from the start of the HTTP request message until (and including) the double CRLF before the body. Set to -1 if the info is not available.
	BodySize    int64     `json:"bodySize"`           // Size of the request body (POST data payload) in bytes. Set to -1 if the info is not available.
	Comment     string    `json:"comment,omitempty"`  // A comment provided by the user or the application.
}

// Response contains detailed info about the response.
type Response struct {
	Status      int64     `json:"status"`            // Response status.
	StatusText  string    `json:"statusText"`        // Response status description.
	HTTPVersion string    `json:"httpVersion"`       // Response HTTP Version.
	Cookies     []*Cookie `json:"cookies"`           // List of cookie objects.
	Headers     []*NVPair `json:"headers"`           // List of header objects.
	Content     *Content  `json:"content"`           // Details about the response body.
	RedirectURL *string   `json:"redirectURL"`       // Redirection target URL from the Location response header.
	HeadersSize int64     `json:"headersSize"`       // Total number of bytes from the start of the HTTP response message until (and including) the double CRLF before the body. Set to -1 if the info is not available.
	BodySize    int64     `json:"bodySize"`          // Size of the received response body in bytes. Set to zero in case of responses coming from the cache (304). Set to -1 if the info is not available.
	Comment     string    `json:"comment,omitempty"` // A comment provided by the user or the application.
}

// Cookie contains list of all cookies (used in [Request] and [Response] objects).
type Cookie struct {
	Name     string `json:"name"`               // The name of the cookie.
	Value    string `json:"value"`              // The cookie value.
	Path     string `json:"path,omitempty"`     // The path pertaining to the cookie.
	Domain   string `json:"domain,omitempty"`   // The host of the cookie.
	Expires  string `json:"expires,omitempty"`  // Cookie expiration time. (ISO 8601 - YYYY-MM-DDThh:mm:ss.sTZD, e.g. 2009-07-24T19:20:30.123+02:00).
	HTTPOnly bool   `json:"httpOnly,omitempty"` // Set to true if the cookie is HTTP only, false otherwise.
	Secure   bool   `json:"secure,omitempty"`   // True if the cookie was transmitted over ssl, false otherwise.
	Comment  string `json:"comment,omitempty"`  // A comment provided by the user or the application.
}

// NVPair describes a name/value pair.
type NVPair struct {
	Name    string `json:"name"`              // Name of the pair.
	Value   string `json:"value"`             // Value of the pair.
	Comment string `json:"comment,omitempty"` // A comment provided by the user or the application.
}

// PostData describes posted data, if any (embedded in [Request] object). Text and params fields are
// mutually exclusive.
type PostData struct {
	MimeType string   `json:"mimeType"`          // Mime type of posted data.
	Params   []*Param `json:"params,omitempty"`  // List of posted parameters (in case of URL encoded parameters).
	Text     string   `json:"text,omitempty"`    // Plain text posted data
	Comment  string   `json:"comment,omitempty"` // A comment provided by the user or the application.
}

// Param list of posted parameters, if any (embedded in [PostData] object).
type Param struct {
	Name        string `json:"name"`                  // Name of a posted parameter.
	Value       string `json:"value,omitempty"`       // Value of a posted parameter or content of a posted file.
	FileName    string `json:"fileName,omitempty"`    // Name of a posted file.
	ContentType string `json:"contentType,omitempty"` // Content type of a posted file.
	Comment     string `json:"comment,omitempty"`     // A comment provided by the user or the application.
}

// Content describes details about response content (embedded in [Response] object).
type Content struct {
	Size        int64  `json:"size"`                  // Length of the returned content in bytes. Should be equal to response.bodySize if there is no compression and bigger when the content has been compressed.
	Compression int64  `json:"compression,omitempty"` // Number of bytes saved. Leave out this field if the information is not available.
	MimeType    string `json:"mimeType"`              // MIME type of the response text (value of the Content-Type response header). The charset attribute of the MIME type is included (if available).
	Text        string `json:"text,omitempty"`        // Response body sent from the server or loaded from the browser cache. This field is populated with textual content only. The text field is either HTTP decoded text or a encoded (e.g. "base64") representation of the response body. Leave out this field if the information is not available.
	Encoding    string `json:"encoding,omitempty"`    // Encoding used for response text field e.g "base64". Leave out this field if the text field is HTTP decoded (decompressed & unchunked), than trans-coded from its original character set into UTF-8.
	Comment     string `json:"comment,omitempty"`     // A comment provided by the user or the application.
}

// Cache contains info about a request coming from browser cache.
type Cache struct {
	BeforeRequest *CacheData `json:"beforeRequest,omitempty"` // State of a cache entry before the request. Leave out this field if the information is not available.
	AfterRequest  *CacheData `json:"afterRequest,omitempty"`  // State of a cache entry after the request. Leave out this field if the information is not available.
	Comment       string     `json:"comment,omitempty"`       // A comment provided by the user or the application.
}

// CacheData describes the cache data for beforeRequest and afterRequest.
type CacheData struct {
	Expires    string `json:"expires,omitempty"` // Expiration time of the cache entry.
	LastAccess string `json:"lastAccess"`        // The last time the cache entry was opened.
	ETag       string `json:"eTag"`              // Etag
	HitCount   int64  `json:"hitCount"`          // The number of times the cache entry has been opened.
	Comment    string `json:"comment,omitempty"` // A comment provided by the user or the application.
}

// Timings describes various phases within request-response round trip. All times are specified in
// milliseconds.
type Timings struct {
	Blocked float64 `json:"blocked,omitempty,omitzero"` // Time spent in a queue waiting for a network connection. Use -1 if the timing does not apply to the current request.
	DNS     float64 `json:"dns,omitempty,omitzero"`     // DNS resolution time. The time required to resolve a host name. Use -1 if the timing does not apply to the current request.
	Connect float64 `json:"connect,omitempty,omitzero"` // Time required to create TCP connection. Use -1 if the timing does not apply to the current request.
	Send    float64 `json:"send"`                       // Time required to send HTTP request to the server.
	Wait    float64 `json:"wait"`                       // Waiting for a response from the server.
	Receive float64 `json:"receive"`                    // Time required to read entire response from the server (or cache).
	Ssl     float64 `json:"ssl,omitempty,omitzero"`     // Time required for SSL/TLS negotiation. If this field is defined then the time is also included in the connect field (to ensure backward compatibility with HAR 1.1). Use -1 if the timing does not apply to the current request.
	Comment string  `json:"comment,omitempty"`          // A comment provided by the user or the application.
}

// Total returns the total time of the request/response round trip. Ignoring -1 values which
// indicate that the timing does not apply to the current request.
func (t *Timings) Total() float64 {
	sum := 0.0
	for _, v := range []float64{t.Blocked, t.DNS, t.Connect, t.Send, t.Wait, t.Receive, t.Ssl} {
		if v > 0 {
			sum += v
		}
	}
	return sum
}

// Save saves the HAR data to a file in JSON format under the specified filename.
// It uses the json.NewEncoder to encode without escaping HTML.
func (h *HAR) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")

	if err := enc.Encode(h); err != nil {
		return err
	}

	return nil
}
