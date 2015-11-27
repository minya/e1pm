package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"
)

var timeout = time.Duration(10 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

var RedirectAttemptedError = errors.New("redirect")

func noRedirect(req *http.Request, via []*http.Request) error {
	return RedirectAttemptedError
}

func main() {
	os.Setenv("HTTP_PROXY", "http://192.168.14.140:8888")
	proxyUrl, _ := url.Parse("https://192.168.14.140:8888")
	runtime.GOMAXPROCS(16)
	transport := http.Transport{
		Dial:            dialTimeout,
		Proxy:           http.ProxyURL(proxyUrl),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	jar := NewJar()
	client := http.Client{
		Transport:     &transport,
		Jar:           jar,
		CheckRedirect: noRedirect,
	}

	pmUrl := "http://www.e1.ru/talk/forum/pm/index.php"
	pmResult, err := client.Get(pmUrl)
	if urlError, ok := err.(*url.Error); ok && urlError.Err == RedirectAttemptedError {
		err = nil
	}
	if nil != err {
		fmt.Printf("error fetch %v\n", pmUrl)
		return
	}
	pmResult.Body.Close()
	if pmResult.StatusCode != 302 {
		fmt.Printf("%v from %v, 302 expected\n", pmResult.StatusCode, pmUrl)
	}

	mye1Location := pmResult.Header.Get("Location")
	fmt.Printf("mye1 location: %v\n", mye1Location)

	respMye1, _ := client.Get(mye1Location)

	checkLocation := respMye1.Header.Get("Location")
	fmt.Printf("Myie: %v %v\n", respMye1.StatusCode, checkLocation)
	fmt.Printf("Check location: %v \n", checkLocation)

	respCheck, _ := client.Get(checkLocation)
	fmt.Printf("Check result: %v\n", respCheck.StatusCode, checkLocation)
	setNgsCookieLocation := respCheck.Header.Get("Location")
	fmt.Printf("Ngs set cookie location: %v\n", setNgsCookieLocation)

	respNgsSetCookie, _ := client.Get(setNgsCookieLocation)
	ngsAfterSetCookieLocation := respNgsSetCookie.Header.Get("Location")
	fmt.Printf("Set ngs cookie result: %v %v\n", respNgsSetCookie.StatusCode, ngsAfterSetCookieLocation)

	respNgsAfterSetCookie, _ := client.Get(ngsAfterSetCookieLocation)
	redirectMye1Location := respNgsAfterSetCookie.Header.Get("Location")
	fmt.Printf("After ngs cookie result: %v %v\n", respNgsAfterSetCookie.StatusCode, redirectMye1Location)

	respMye1Redirected, _ := client.Get(redirectMye1Location)
	fmt.Printf("Redirected mye1 result: %v\n", respMye1Redirected.StatusCode)

	loginUrl := "https://passport.ngs.ru/e1/login/?redirect_path=http%3A%2F%2Fwww.e1.ru%2Ftalk%2Fforum%2Fpm%2Findex.php"
	values := url.Values{
		"sub":      {"login"},
		"key":      {""},
		"email":    {"minya.drel@gmail.com"},
		"password": {""}}

	respLogin, errLogin := client.PostForm(loginUrl, values)
	if urlError, ok := errLogin.(*url.Error); ok && urlError.Err == RedirectAttemptedError {
		errLogin = nil
	}

	if nil != errLogin {
		fmt.Printf("Error login: %v\n", errLogin)
		return
	}

	if 302 != respLogin.StatusCode {
		fmt.Printf("302 expected at login, but %v taken\n", errLogin)
		return
	}

	redirectUrl := respLogin.Header.Get("Location")
	fmt.Printf("redirect url: %v \n", redirectUrl)

	respRedirect, errRedirect := client.Get(redirectUrl)
	if urlError, ok := errRedirect.(*url.Error); ok && urlError.Err == RedirectAttemptedError {
		errRedirect = nil
	}

	if nil != errRedirect {
		fmt.Printf("err on redirect %v\n", errRedirect)
		return
	}

	nextRedirectUrl := respRedirect.Header.Get("Location")
	fmt.Printf("%v %v\n", respRedirect.StatusCode, nextRedirectUrl)

	respPm, errPm := client.Get(nextRedirectUrl)

	if nil != errPm {
		fmt.Printf("Error get pm page after authentication: %v\n", errPm)
		return
	}

	body, errReadBody := ioutil.ReadAll(respPm.Body)
	if nil != errReadBody {
	}

	html := string(body)
	fmt.Printf(html)
}
