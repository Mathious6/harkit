package main

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/Mathious6/harkit/harhandler"
	"github.com/Mathious6/httpkit"
	http "github.com/bogdanfinn/fhttp"
)

const (
	URL        = "https://httpbin.org"
	PROXY_HOST = "host.docker.internal"
	PROXY_PORT = "8888"
)

func main() {
	handler := harhandler.NewHandler()

	opts := []httpkit.HttpClientOption{
		httpkit.WithCookieJar(httpkit.NewCookieJar()),
		httpkit.WithNotFollowRedirects(),
	}

	if isProxyRunning(net.JoinHostPort(PROXY_HOST, PROXY_PORT), 100*time.Millisecond) {
		opts = append(opts, httpkit.WithCharlesProxy(PROXY_HOST, PROXY_PORT))
		fmt.Println("Using Charles proxy.")
	} else {
		fmt.Println("Charles proxy not running, using direct connection.")
	}

	client, err := httpkit.NewHttpClient(httpkit.NewNoopLogger(), opts...)
	if err != nil {
		panic(err)
	}

	sendGetRequestWithQueryParams(handler, client)
	// sendGetRequestWithSetCookies(handler, client)
	// sendPostRequestWithForm(handler, client)
	// sendPostRequestWithJSON(handler, client)

	handler.Save("example.har")
}

func sendGetRequestWithQueryParams(handler *harhandler.HARHandler, client httpkit.HttpClient) {
	req, _ := http.NewRequest(http.MethodGet, URL+"/get?name=pierre&role=developer", nil)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "httpbin.org")
	req.Header.Add("User-Agent", "harkit-example")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")

	req.Header.Add(http.HeaderOrderKey, "accept")
	req.Header.Add(http.HeaderOrderKey, "host")
	req.Header.Add(http.HeaderOrderKey, "user-agent")
	req.Header.Add(http.HeaderOrderKey, "accept-encoding")

	entry := harhandler.NewEntry()
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	_ = entry.AddEntry(req, resp)

	handler.AddEntry(entry)

	fmt.Println("Parameters sent.")
}

func sendGetRequestWithSetCookies(handler *harhandler.HARHandler, client httpkit.HttpClient) {
	req, _ := http.NewRequest(http.MethodGet, URL+"/cookies/set?name=pierre&role=developer", nil)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "httpbin.org")
	req.Header.Add("User-Agent", "harkit-example")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")

	req.Header.Add(http.HeaderOrderKey, "accept")
	req.Header.Add(http.HeaderOrderKey, "host")
	req.Header.Add(http.HeaderOrderKey, "user-agent")
	req.Header.Add(http.HeaderOrderKey, "accept-encoding")

	entry := harhandler.NewEntry()
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	_ = entry.AddEntry(req, resp)

	handler.AddEntry(entry)

	fmt.Println("Cookies set.")
}

func sendPostRequestWithForm(handler *harhandler.HARHandler, client httpkit.HttpClient) {
	form := url.Values{}
	form.Set("name", "Pierre")
	form.Set("role", "developer")
	body := strings.NewReader(form.Encode())

	req, _ := http.NewRequest(http.MethodPost, URL+"/post", body)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Host", "httpbin.org")
	req.Header.Add("User-Agent", "harkit-example")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")

	req.AddCookie(&http.Cookie{Name: "example", Value: "cookie"})

	req.Header.Add(http.HeaderOrderKey, "accept")
	req.Header.Add(http.HeaderOrderKey, "content-length")
	req.Header.Add(http.HeaderOrderKey, "content-type")
	req.Header.Add(http.HeaderOrderKey, "cookie")
	req.Header.Add(http.HeaderOrderKey, "host")
	req.Header.Add(http.HeaderOrderKey, "user-agent")
	req.Header.Add(http.HeaderOrderKey, "accept-encoding")

	entry := harhandler.NewEntry()
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	_ = entry.AddEntry(req, resp)

	handler.AddEntry(entry)

	fmt.Println("Form URL-encoded request sent.")
}

func sendPostRequestWithJSON(handler *harhandler.HARHandler, client httpkit.HttpClient) {
	jsonBody := `{"name":"Pierre","role":"developer"}`
	body := strings.NewReader(jsonBody)

	req, _ := http.NewRequest(http.MethodPost, URL+"/post", body)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Host", "httpbin.org")
	req.Header.Add("User-Agent", "harkit-example")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")

	req.AddCookie(&http.Cookie{Name: "example", Value: "cookie"})

	req.Header.Add(http.HeaderOrderKey, "accept")
	req.Header.Add(http.HeaderOrderKey, "content-length")
	req.Header.Add(http.HeaderOrderKey, "content-type")
	req.Header.Add(http.HeaderOrderKey, "cookie")
	req.Header.Add(http.HeaderOrderKey, "host")
	req.Header.Add(http.HeaderOrderKey, "user-agent")
	req.Header.Add(http.HeaderOrderKey, "accept-encoding")

	entry := harhandler.NewEntry()
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	_ = entry.AddEntry(req, resp)

	handler.AddEntry(entry)

	fmt.Println("JSON request sent.")
}

// isProxyRunning checks if a proxy is running on the given address and port.
func isProxyRunning(address string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
