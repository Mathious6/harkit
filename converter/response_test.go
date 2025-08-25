package converter_test

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/Mathious6/harkit/converter"
	http "github.com/bogdanfinn/fhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	RESP_OK_CODE    = http.StatusOK
	RESP_FOUND_CODE = http.StatusFound
	RESP_PROTOCOL   = "HTTP/1.1"

	RESP_HEADER_NAME  = "Name"
	RESP_HEADER_VALUE = "value"

	RESP_COOKIE_NAME     = "name"
	RESP_COOKIE_VALUE    = "value"
	RESP_COOKIE_PATH     = "/"
	RESP_COOKIE_DOMAIN   = "example.com"
	RESP_COOKIE_EXPIRES  = "Mon, 12 May 2025 00:00:00 GMT"
	RESP_COOKIE_HTTPONLY = true
	RESP_COOKIE_SECURE   = true

	RESP_BODY_TEXT                 = "response"
	RESP_CONTENT_TYPE              = "text/plain"
	RESP_JSON_CONTENT_LENGTH_VALUE = "8"

	RESP_LOCATION = "https://example.com/redirect"
)

func TestConverter_GivenNilResponse_WhenConvertingHTTPResponse_ThenErrorShouldBeReturned(t *testing.T) {
	resp := (*http.Response)(nil)

	result, err := converter.FromHTTPResponse(resp)

	assert.Error(t, err, "Error should be returned when response is nil")
	assert.Nil(t, result, "HAR should be nil when response is nil")
}

func TestConverter_GivenStatusCode_WhenConvertingHTTPResponse_ThenStatusShouldBeCorrect(t *testing.T) {
	resp := createResponse(t, RESP_OK_CODE, nil, "")

	result, err := converter.FromHTTPResponse(resp)
	require.NoError(t, err)

	assert.Equal(t, int64(resp.StatusCode), result.Status, "HAR status <> response status")
}

func TestConverter_GivenStatusText_WhenConvertingHTTPResponse_ThenStatusTextShouldBeCorrect(t *testing.T) {
	resp := createResponse(t, RESP_OK_CODE, nil, "")

	result, err := converter.FromHTTPResponse(resp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusText(resp.StatusCode), result.StatusText, "HAR status text <> response status text")
}

func TestConverter_GivenProtocol_WhenConvertingHTTPResponse_ThenProtocolShouldBeCorrect(t *testing.T) {
	resp := createResponse(t, RESP_OK_CODE, nil, "")

	result, err := converter.FromHTTPResponse(resp)
	require.NoError(t, err)

	assert.Equal(t, resp.Proto, result.HTTPVersion, "HAR protocol <> response protocol")
}

func TestConverter_GivenCookies_WhenConvertingHTTPResponse_ThenCookiesShouldBeCorrect(t *testing.T) {
	resp := createResponse(t, RESP_OK_CODE, nil, "")

	result, err := converter.FromHTTPResponse(resp)
	require.NoError(t, err)

	exceptedExpires, _ := time.Parse(time.RFC1123, RESP_COOKIE_EXPIRES)
	exceptedExpiresStr := exceptedExpires.Format(time.RFC3339Nano)
	assert.Len(t, result.Cookies, 1, "HAR should contain 1 cookie")
	assert.Equal(t, RESP_COOKIE_NAME, result.Cookies[0].Name, "HAR cookie name <> response cookie name")
	assert.Equal(t, RESP_COOKIE_VALUE, result.Cookies[0].Value, "HAR cookie value <> response cookie value")
	assert.Equal(t, RESP_COOKIE_PATH, result.Cookies[0].Path, "HAR cookie path <> response cookie path")
	assert.Equal(t, RESP_COOKIE_DOMAIN, result.Cookies[0].Domain, "HAR cookie domain <> response cookie domain")
	assert.Equal(t, exceptedExpiresStr, result.Cookies[0].Expires, "HAR cookie expires <> response cookie expires")
	assert.Equal(t, RESP_COOKIE_HTTPONLY, result.Cookies[0].HTTPOnly, "HAR cookie httpOnly <> response cookie httpOnly")
	assert.Equal(t, RESP_COOKIE_SECURE, result.Cookies[0].Secure, "HAR cookie secure <> response cookie secure")
}

func TestConverter_GivenHeaders_WhenConvertingHTTPResponse_ThenHeadersShouldBeCorrect(t *testing.T) {
	resp := createResponse(t, RESP_OK_CODE, nil, "")

	result, err := converter.FromHTTPResponse(resp)
	require.NoError(t, err)

	assert.Len(t, result.Headers, 2, "HAR should contain 2 header")
	assert.Equal(t, RESP_HEADER_NAME, result.Headers[0].Name, "HAR header name <> response header name")
	assert.Equal(t, RESP_HEADER_VALUE, result.Headers[0].Value, "HAR header value <> response header value")
}

func TestConverter_GivenBody_WhenConvertingHTTPResponse_ThenContentShouldBeCorrect(t *testing.T) {
	resp := createResponse(t, RESP_OK_CODE, bytes.NewBufferString(RESP_BODY_TEXT), RESP_CONTENT_TYPE)

	result, err := converter.FromHTTPResponse(resp)
	require.NoError(t, err)

	assert.Equal(t, int64(len(RESP_BODY_TEXT)), result.Content.Size, "HAR content size <> response body size")
	assert.Equal(t, RESP_CONTENT_TYPE, result.Content.MimeType, "HAR content mime type <> response content mime type")
	assert.Equal(t, RESP_BODY_TEXT, result.Content.Text, "HAR content text <> response body text")
}

func TestConverter_GivenRedirect_WhenConvertingHTTPResponse_ThenRedirectURLShouldBeCorrect(t *testing.T) {
	resp := createResponse(t, RESP_FOUND_CODE, nil, "")

	result, err := converter.FromHTTPResponse(resp)
	require.NoError(t, err)

	assert.Equal(t, RESP_LOCATION, *result.RedirectURL, "HAR redirect URL <> response location header")
}

func TestConverter_GivenBody_WhenConvertingHTTPResponse_ThenBodySizeShouldBeCorrect(t *testing.T) {
	resp := createResponse(t, RESP_OK_CODE, bytes.NewBufferString(RESP_BODY_TEXT), RESP_CONTENT_TYPE)

	result, err := converter.FromHTTPResponse(resp)
	require.NoError(t, err)

	assert.Equal(t, int64(len(RESP_BODY_TEXT)), result.BodySize, "HAR body size <> response body size")
}

func TestConverter_GivenEmptyBody_WhenConvertingHTTPResponse_ThenContentLengthShouldNotBeSet(t *testing.T) {
	resp := createResponse(t, RESP_OK_CODE, bytes.NewBufferString(""), RESP_CONTENT_TYPE)

	result, err := converter.FromHTTPResponse(resp)
	require.NoError(t, err)

	assert.Len(t, result.Headers, 3, "HAR should contain 3 header")
}

func TestConverter_GivenBody_WhenConvertingHTTPResponse_ThenContentLengthShouldBeSet(t *testing.T) {
	resp := createResponse(t, RESP_OK_CODE, bytes.NewBufferString(RESP_BODY_TEXT), RESP_CONTENT_TYPE)

	result, err := converter.FromHTTPResponse(resp)
	require.NoError(t, err)

	assert.Len(t, result.Headers, 4, "HAR should contain 4 header")
	assert.Equal(t, RESP_JSON_CONTENT_LENGTH_VALUE, result.Headers[3].Value, "HAR content length <> response content length")
}

func createResponse(t *testing.T, statusCode int, body io.Reader, contentType string) *http.Response {
	var buf []byte
	var err error
	if body != nil {
		buf, err = io.ReadAll(body)
		require.NoError(t, err)
	}

	resp := &http.Response{
		StatusCode:    statusCode,
		Status:        http.StatusText(statusCode),
		Proto:         RESP_PROTOCOL,
		Header:        make(http.Header),
		ContentLength: int64(len(buf)),
		Body:          io.NopCloser(bytes.NewBuffer(buf)),
	}

	cookie := RESP_COOKIE_NAME + "=" + RESP_COOKIE_VALUE
	cookie += ";path=" + RESP_COOKIE_PATH
	cookie += ";domain=" + RESP_COOKIE_DOMAIN
	cookie += ";expires=" + RESP_COOKIE_EXPIRES
	cookie += ";httponly"
	cookie += ";secure"

	resp.Header.Append(RESP_HEADER_NAME, RESP_HEADER_VALUE)
	resp.Header.Append(converter.SetCookieKey, cookie)

	if contentType != "" {
		resp.Header.Append(converter.ContentTypeKey, contentType)
	}

	if len(buf) > 0 {
		resp.Header.Append(converter.ContentLengthKey, RESP_JSON_CONTENT_LENGTH_VALUE)
	}

	if statusCode >= 300 && statusCode < 400 {
		resp.Header.Append(converter.LocationKey, RESP_LOCATION)
	}

	return resp
}
