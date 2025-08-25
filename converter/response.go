package converter

import (
	"bytes"
	"errors"
	"io"
	"strconv"

	"github.com/Mathious6/harkit/harfile"
	http "github.com/bogdanfinn/fhttp"
)

func FromHTTPResponse(resp *http.Response) (*harfile.Response, error) {
	if resp == nil {
		return nil, errors.New("response cannot be nil")
	}

	content, err := buildResponseContent(resp)
	if err != nil {
		return nil, err
	}

	protocolHeader := handleResponseProtocolHeader(resp.Proto, resp.StatusCode)
	headers := convertHeaders(resp.Header, resp.ContentLength)

	return &harfile.Response{
		Status:      int64(resp.StatusCode),
		StatusText:  http.StatusText(resp.StatusCode),
		HTTPVersion: resp.Proto,
		Cookies:     convertCookies(resp.Cookies()),
		Headers:     append(protocolHeader, headers...),
		Content:     content,
		RedirectURL: locateRedirectURL(resp),
		HeadersSize: -1,
		BodySize:    content.Size,
	}, nil
}

func handleResponseProtocolHeader(proto string, status int) []*harfile.NVPair {
	if proto == "HTTP/2.0" {
		return []*harfile.NVPair{
			{Name: ":status", Value: strconv.Itoa(status)},
		}
	} else {
		return []*harfile.NVPair{}
	}
}

func locateRedirectURL(resp *http.Response) *string {
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		if loc, err := resp.Location(); err == nil {
			url := loc.String()
			return &url
		}
	}
	return nil
}

func buildResponseContent(resp *http.Response) (*harfile.Content, error) {
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(buf))

	return &harfile.Content{
		Size:        int64(len(buf)),
		Compression: 0,
		MimeType:    resp.Header.Get(ContentTypeKey),
		Text:        string(buf),
		Encoding:    "",
	}, nil
}
