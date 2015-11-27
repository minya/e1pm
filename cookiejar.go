package main

import (
	"net/http"
	"net/url"
	"sync"
)

type Jar struct {
	lk      sync.Mutex
	cookies map[string]map[string]*http.Cookie
}

func NewJar() *Jar {
	jar := new(Jar)
	jar.cookies = make(map[string]map[string]*http.Cookie)
	return jar
}

// SetCookies handles the receipt of the cookies in a reply for the
// given URL.  It may or may not choose to save the cookies, depending
// on the jar's policy and implementation.
func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.lk.Lock()
	if jar.cookies[u.Host] == nil {
		jar.cookies[u.Host] = make(map[string]*http.Cookie)
	}

	for _, cookie := range cookies {
		jar.cookies[u.Host][cookie.Name] = cookie
	}
	jar.lk.Unlock()
}

// Cookies returns the cookies to send in a request for the given URL.
// It is up to the implementation to honor the standard cookie use
// restrictions such as in RFC 6265.
func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	r := make([]*http.Cookie, len(jar.cookies[u.Host]))
	idx := 0
	for _, c := range jar.cookies[u.Host] {
		r[idx] = c
		idx++
	}
	return r
}
