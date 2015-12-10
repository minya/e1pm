package main

import (
	"net"
	"net/http"
	"time"
)

var timeout = time.Duration(10 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

//proxyUrl, _ := url.Parse("https://192.168.14.140:8888")
func DefaultTransport() *http.Transport {
	transport := new(http.Transport)
	transport.Dial = dialTimeout
	//transport.Proxy = http.ProxyURL(proxyUrl),
	//transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true},
	return transport
}
