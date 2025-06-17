package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Mathious6/harkit/harhandler"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

const URL = "https://httpbin.org"

func main() {
	handler := harhandler.NewHandler()

	client, _ := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithCookieJar(tls_client.NewCookieJar()),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCharlesProxy("host.docker.internal", "8888"),
	)

	sendGetRequestWithQueryParams(handler, client)
	sendGetRequestWithSetCookies(handler, client)
	sendPostRequestWithForm(handler, client)
	sendPostRequestWithJSON(handler, client)

	handler.Save("example.har")
}

func sendGetRequestWithQueryParams(handler *harhandler.HARHandler, client tls_client.HttpClient) {
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

func sendGetRequestWithSetCookies(handler *harhandler.HARHandler, client tls_client.HttpClient) {
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

func sendPostRequestWithForm(handler *harhandler.HARHandler, client tls_client.HttpClient) {
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

func sendPostRequestWithJSON(handler *harhandler.HARHandler, client tls_client.HttpClient) {
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
