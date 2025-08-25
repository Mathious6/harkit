package converter

import (
	"errors"
	"io"
	"net/url"
	"strings"

	"github.com/Mathious6/harkit/harfile"
	http "github.com/bogdanfinn/fhttp"
)

const (
	applicationXWWWFormURLEncoded = "application/x-www-form-urlencoded"
	multipartFormData             = "multipart/form-data"
	maxMultipartFormDataSize      = 32 << 20 // 32 MB limit

	methodKey    = ":method"
	authorityKey = ":authority"
	schemeKey    = ":scheme"
	pathKey      = ":path"

	hostKey = "Host"
)

func FromHTTPRequest(req *http.Request) (*harfile.Request, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	reqProto := DefaultRequestHTTPVersion // WARNING: req.Proto is not always accurate so we force it.

	protocolHeader := handleRequestProtocolHeader(reqProto, req.Method, *req.URL)
	headers := convertHeaders(req.Header, req.ContentLength)

	postData, err := extractRequestPostData(req)
	if err != nil {
		return nil, err
	}

	return &harfile.Request{
		Method:      req.Method,
		URL:         req.URL.String(),
		HTTPVersion: reqProto,
		Cookies:     convertCookies(req.Cookies()),
		Headers:     append(protocolHeader, headers...),
		QueryString: convertRequestQueryParams(req.URL),
		PostData:    postData,
		HeadersSize: -1,
		BodySize:    req.ContentLength,
	}, nil
}

func handleRequestProtocolHeader(proto string, method string, url url.URL) []*harfile.NVPair {
	if proto == "HTTP/2.0" {
		return []*harfile.NVPair{
			{Name: methodKey, Value: method},
			{Name: authorityKey, Value: url.Host},
			{Name: schemeKey, Value: url.Scheme},
			{Name: pathKey, Value: url.RequestURI()},
		}
	} else {
		return []*harfile.NVPair{
			{Name: hostKey, Value: url.Host},
		}
	}
}

func convertRequestQueryParams(u *url.URL) []*harfile.NVPair {
	result := make([]*harfile.NVPair, 0)

	for key, values := range u.Query() {
		for _, value := range values {
			result = append(result, &harfile.NVPair{Name: key, Value: value})
		}
	}

	return result
}

func extractRequestPostData(req *http.Request) (*harfile.PostData, error) {
	if req.Body == nil || req.ContentLength == 0 {
		return nil, nil
	}

	body, err := req.GetBody()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	bodyText, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	mimeType := req.Header.Get(ContentTypeKey)
	postData := &harfile.PostData{MimeType: mimeType}

	if strings.HasPrefix(mimeType, applicationXWWWFormURLEncoded) {
		pairs := strings.SplitSeq(string(bodyText), "&")

		for pair := range pairs {
			nv := strings.SplitN(pair, "=", 2)
			if len(nv) == 2 {
				name, value := nv[0], nv[1]
				postData.Params = append(postData.Params, &harfile.Param{Name: name, Value: value})
			}
		}

		return postData, nil
	}

	if strings.HasPrefix(mimeType, multipartFormData) {
		err := req.ParseMultipartForm(maxMultipartFormDataSize)
		if err != nil {
			return nil, err
		}

		for name, values := range req.MultipartForm.Value {
			for _, value := range values {
				postData.Params = append(postData.Params, &harfile.Param{Name: name, Value: value})
			}
		}

		for name, files := range req.MultipartForm.File {
			for _, fileHeader := range files {
				file, err := fileHeader.Open()
				if err != nil {
					return nil, err
				}
				defer file.Close()

				content, err := io.ReadAll(file)
				if err != nil {
					return nil, err
				}

				postData.Params = append(postData.Params, &harfile.Param{
					Name:        name,
					FileName:    fileHeader.Filename,
					ContentType: fileHeader.Header.Get(ContentTypeKey),
					Value:       string(content),
				})
			}
		}

		return postData, nil
	}

	postData.Text = string(bodyText)
	return postData, nil
}
