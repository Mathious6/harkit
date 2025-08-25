package converter_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/Mathious6/harkit/converter"
	http "github.com/bogdanfinn/fhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	REQ_METHOD                 = http.MethodPost
	REQ_URL                    = "https://example.com/api?foo=bar"
	REQ_URI                    = "/api?foo=bar"
	REQ_PROTOCOL               = "HTTP/2.0"
	REQ_PROTOCOL_HEADERS_COUNT = 4

	REQ_HEADER1_NAME  = "Name1"
	REQ_HEADER1_VALUE = "value1"
	REQ_HEADER2_NAME  = "Name2"
	REQ_HEADER2_VALUE = "value2"

	REQ_COOKIE_NAME  = "name"
	REQ_COOKIE_VALUE = "value"

	REQ_URL_CONTENT_TYPE = "application/x-www-form-urlencoded"
	REQ_BODY_URL         = "foo=bar"

	REQ_JSON_CONTENT_TYPE         = "application/json"
	REQ_BODY_JSON                 = `{"foo":"bar"}`
	REQ_JSON_CONTENT_LENGTH_VALUE = "13"

	REQ_PART1_NAME         = "name1"
	REQ_PART1_VALUE        = "value1"
	REQ_PART2_NAME         = "file"
	REQ_PART2_VALUE        = "content"
	REQ_PART2_FILENAME     = "test.txt"
	REQ_PART2_CONTENT_TYPE = "application/octet-stream"
)

func TestConverter_GivenNilRequest_WhenConvertingHTTPRequest_ThenErrorShouldBeReturned(t *testing.T) {
	req := (*http.Request)(nil)

	result, err := converter.FromHTTPRequest(req)

	assert.Error(t, err, "Error should be returned when request is nil")
	assert.Nil(t, result, "HAR should be nil when request is nil")
}

func TestConverter_GivenMethod_WhenConvertingHTTPRequest_ThenMethodShouldBeCorrect(t *testing.T) {
	req := createRequest(t, nil, "")

	result, err := converter.FromHTTPRequest(req)
	require.NoError(t, err)

	assert.Equal(t, REQ_METHOD, result.Method, "HAR method <> request method")
}

func TestConverter_GivenURL_WhenConvertingHTTPRequest_ThenURLShouldBeCorrect(t *testing.T) {
	req := createRequest(t, nil, "")

	result, err := converter.FromHTTPRequest(req)
	require.NoError(t, err)

	assert.Equal(t, REQ_URL, result.URL, "HAR URL <> request URL")
}

func TestConverter_GivenProtocol_WhenConvertingHTTPRequest_ThenProtocolShouldBeCorrect(t *testing.T) {
	req := createRequest(t, nil, "")

	result, err := converter.FromHTTPRequest(req)
	require.NoError(t, err)

	// assert.Equal(t, REQ_PROTOCOL, result.HTTPVersion, "HAR protocol <> request protocol")
	assert.Equal(t, "HTTP/2.0", result.HTTPVersion, "HAR protocol <> request protocol") // WARNING: we force it.
}

func TestConverter_GivenCookies_WhenConvertingHTTPRequest_ThenCookiesShouldBeCorrect(t *testing.T) {
	req := createRequest(t, nil, "")

	result, err := converter.FromHTTPRequest(req)
	require.NoError(t, err)

	assert.Len(t, result.Cookies, 1, "HAR should contain 1 cookie")
	assert.Equal(t, REQ_COOKIE_NAME, result.Cookies[0].Name, "HAR cookie name <> request cookie name")
	assert.Equal(t, REQ_COOKIE_VALUE, result.Cookies[0].Value, "HAR cookie value <> request cookie value")
}

func TestConverter_GivenHeaders_WhenConvertingHTTPRequest_ThenHeadersShouldBeCorrectAndOrdered(t *testing.T) {
	req := createRequest(t, nil, "")

	result, err := converter.FromHTTPRequest(req)
	require.NoError(t, err)

	assert.Len(t, result.Headers, REQ_PROTOCOL_HEADERS_COUNT+3, "HAR should contain 3 headers")
	assert.Equal(t, REQ_HEADER2_NAME, result.Headers[REQ_PROTOCOL_HEADERS_COUNT+0].Name, "HAR header name <> request header name")
	assert.Equal(t, REQ_HEADER2_VALUE, result.Headers[REQ_PROTOCOL_HEADERS_COUNT+0].Value, "HAR header value <> request header value")
	assert.Equal(t, REQ_HEADER1_NAME, result.Headers[REQ_PROTOCOL_HEADERS_COUNT+1].Name, "HAR header name <> request header name")
	assert.Equal(t, REQ_HEADER1_VALUE, result.Headers[REQ_PROTOCOL_HEADERS_COUNT+1].Value, "HAR header value <> request header value")
}

func TestConverter_GivenURLWithQueryString_WhenConvertingHTTPRequest_ThenQueryStringShouldBeCorrect(t *testing.T) {
	req := createRequest(t, nil, "")

	result, err := converter.FromHTTPRequest(req)
	require.NoError(t, err)

	assert.Len(t, result.QueryString, 1, "HAR should contain 1 query string parameters")
	assert.Equal(t, "foo", result.QueryString[0].Name, "HAR query string name <> request query string name")
	assert.Equal(t, "bar", result.QueryString[0].Value, "HAR query string value <> request query string value")
}

func TestConverter_GivenEmptyBody_WhenConvertingHTTPRequest_ThenContentLengthShouldNotBeSet(t *testing.T) {
	req := createRequest(t, nil, "")

	result, err := converter.FromHTTPRequest(req)
	require.NoError(t, err)

	assert.NotEqual(t, converter.ContentLengthKey, result.Headers[2].Name, "HAR content length header value should not be set")
}

func TestConverter_GivenURLEncodedBody_WhenConvertingHTTPRequest_ThenPostDataShouldBeCorrect(t *testing.T) {
	req := createRequest(t, strings.NewReader(REQ_BODY_URL), REQ_URL_CONTENT_TYPE)

	result, err := converter.FromHTTPRequest(req)
	require.NoError(t, err)

	assert.Len(t, result.PostData.Params, 1, "HAR should contain 1 post data parameters")
	assert.Empty(t, result.PostData.Text, "HAR should have no post data text")
	assert.Equal(t, "foo", result.PostData.Params[0].Name, "HAR post data name <> request post data name")
	assert.Equal(t, "bar", result.PostData.Params[0].Value, "HAR post data value <> request post data value")
	assert.Equal(t, REQ_URL_CONTENT_TYPE, result.PostData.MimeType, "HAR post data mime type <> request post data mime type")
}

func TestConverter_GivenJSONBody_WhenConvertingHTTPRequest_ThenPostDataShouldBeCorrect(t *testing.T) {
	req := createRequest(t, strings.NewReader(REQ_BODY_JSON), REQ_JSON_CONTENT_TYPE)

	result, err := converter.FromHTTPRequest(req)
	require.NoError(t, err)

	assert.Len(t, result.PostData.Params, 0, "HAR should contain 0 post data parameters")
	assert.NotEmpty(t, result.PostData.Text, "HAR should have post data text")
	assert.Equal(t, REQ_BODY_JSON, result.PostData.Text, "HAR post data text <> request post data text")
	assert.Equal(t, REQ_JSON_CONTENT_TYPE, result.PostData.MimeType, "HAR post data mime type <> request post data mime type")
}

func TestConverter_GivenJSONBody_WhenConvertingHTTPRequest_ThenContentLengthShouldBeSet(t *testing.T) {
	req := createRequest(t, strings.NewReader(REQ_BODY_JSON), REQ_JSON_CONTENT_TYPE)

	result, err := converter.FromHTTPRequest(req)
	require.NoError(t, err)

	assert.Equal(t, converter.ContentLengthKey, result.Headers[REQ_PROTOCOL_HEADERS_COUNT+2].Name, "HAR content length header name <> request content length header name")
	assert.Equal(t, REQ_JSON_CONTENT_LENGTH_VALUE, result.Headers[REQ_PROTOCOL_HEADERS_COUNT+2].Value, "HAR content length header value <> request content length header value")
}

func TestConverter_GivenMultipartBody_WhenConvertingHTTPRequest_ThenPostDataShouldBeCorrect(t *testing.T) {
	body, contentType := createMultipartBody()
	req := createRequest(t, &body, contentType)

	result, err := converter.FromHTTPRequest(req)
	require.NoError(t, err)

	assert.Len(t, result.PostData.Params, 2, "HAR should contain 2 post data parameters")
	assert.Empty(t, result.PostData.Text, "HAR should have no post data text")
	assert.Equal(t, REQ_PART1_NAME, result.PostData.Params[0].Name, "HAR post data name <> request post data name")
	assert.Equal(t, REQ_PART1_VALUE, result.PostData.Params[0].Value, "HAR post data value <> request post data value")
	assert.Equal(t, REQ_PART2_NAME, result.PostData.Params[1].Name, "HAR post data name <> request post data name")
	assert.Equal(t, REQ_PART2_VALUE, result.PostData.Params[1].Value, "HAR post data value <> request post data value")
	assert.Equal(t, REQ_PART2_FILENAME, result.PostData.Params[1].FileName, "HAR post data filename <> request post data filename")
	assert.Equal(t, REQ_PART2_CONTENT_TYPE, result.PostData.Params[1].ContentType, "HAR post data content type <> request post data content type")
	assert.Equal(t, contentType, result.PostData.MimeType, "HAR post data mime type <> request post data mime type")
}

func TestConverter_GivenHeaders_WhenConvertingHTTPRequest_ThenHeadersSizeShouldBeCorrect(t *testing.T) {
	req := createRequest(t, nil, "")

	result, err := converter.FromHTTPRequest(req)
	require.NoError(t, err)

	// assert.Equal(t, computeHeadersSize(), result.HeadersSize, "HAR header size <> request header size")
	assert.Equal(t, int64(-1), result.HeadersSize, "HAR header size <> request header size") // WARNING: we force it.
}

func TestConverter_GivenBody_WhenConvertingHTTPRequest_ThenBodySizeShouldBeCorrect(t *testing.T) {
	req := createRequest(t, strings.NewReader(REQ_BODY_URL), REQ_URL_CONTENT_TYPE)

	result, err := converter.FromHTTPRequest(req)
	require.NoError(t, err)

	expectedSize := int64(len(REQ_BODY_URL))
	assert.Equal(t, expectedSize, result.BodySize, "HAR body size <> request body size")
}

// createRequest creates and returns a new HTTP request with the specified body and content type.
// It sets various headers including cookies, custom headers, and content type if provided.
// The function also ensures the request protocol is set and validates the request creation.
func createRequest(t *testing.T, body io.Reader, contentType string) *http.Request {
	req, err := http.NewRequest(REQ_METHOD, REQ_URL, body)
	require.NoError(t, err)

	req.Proto = REQ_PROTOCOL

	req.Header.Add(converter.CookieKey, REQ_COOKIE_NAME+"="+REQ_COOKIE_VALUE)
	req.Header.Add(REQ_HEADER1_NAME, REQ_HEADER1_VALUE)
	req.Header.Add(REQ_HEADER2_NAME, REQ_HEADER2_VALUE)

	req.Header.Add(http.HeaderOrderKey, REQ_HEADER2_NAME)
	req.Header.Add(http.HeaderOrderKey, REQ_HEADER1_NAME)
	req.Header.Add(http.HeaderOrderKey, converter.ContentLengthKey)

	if contentType != "" {
		req.Header.Add(converter.ContentTypeKey, contentType)
	}

	return req
}

// createMultipartBody constructs a multipart HTTP request body with predefined fields and file content.
// It returns the body as a bytes.Buffer and the corresponding Content-Type header value.
func createMultipartBody() (body bytes.Buffer, contentType string) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	_ = writer.WriteField(REQ_PART1_NAME, REQ_PART1_VALUE)

	fileWriter, _ := writer.CreateFormFile(REQ_PART2_NAME, REQ_PART2_FILENAME)
	_, _ = fileWriter.Write([]byte(REQ_PART2_VALUE))

	writer.Close()

	return buf, writer.FormDataContentType()
}
