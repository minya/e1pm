package main

import (
	//"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
)

var RedirectAttemptedError = errors.New("redirect")

func noRedirect(req *http.Request, via []*http.Request) error {
	return RedirectAttemptedError
}

func main() {
	runtime.GOMAXPROCS(16)

	jar := NewJar()
	transport := DefaultTransport()
	client := http.Client{
		Transport:     transport,
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

	settingsBin, settingsErr := ioutil.ReadFile("config.json")
	if settingsErr != nil {
		fmt.Printf("Can't read settings\n", settingsErr)
		return
	}

	var settings Settings
	//var f interface{}
	errSettings := json.Unmarshal(settingsBin, &settings)
	fmt.Printf("Settings obtained: %v\n", settings)
	if nil != errSettings {
		fmt.Printf("Can't read settings: %v\n", errSettings)
		return
	}

	loginUrl := "https://passport.ngs.ru/e1/login/?redirect_path=http%3A%2F%2Fwww.e1.ru%2Ftalk%2Fforum%2Fpm%2F"
	values := url.Values{
		"sub":      {"login"},
		"key":      {""},
		"email":    {settings.Credentials.Email},
		"password": {settings.Credentials.Password},
	}

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

	tr := transform.NewReader(respPm.Body, charmap.Windows1251.NewDecoder())

	body, errReadBody := ioutil.ReadAll(tr)

	if nil != errReadBody {
		fmt.Errorf("Can't read body: %v\n", errReadBody)
		return
	}

	html := string(body)

	regex, _ := regexp.Compile("<span class=\"text_orange\"><strong>Новое!</strong></span><br><span class=\"small_gray\">(.*)</span>")

	match := regex.FindStringSubmatch(html)
	if len(match) < 2 {
		fmt.Println("No new messages matched")
		return
	}

	dateOfLastPm := match[1]
	dateOfLastSeenPmBin, _ := ioutil.ReadFile("lastseen.txt")
	if dateOfLastPm != string(dateOfLastSeenPmBin) {
		sendRes, sendErr := SendMessage(settings.Pushover.Token, settings.Pushover.User, "New PM on e1.ru")
		if nil != sendErr {
			fmt.Printf("Can't send push: %v\n", sendErr)
		} else {
			fmt.Printf("Push sent: %v\n", sendRes)
		}
		fmt.Printf("New! %v\n", dateOfLastPm)
		ioutil.WriteFile("lastseen.txt", []byte(dateOfLastPm), 0660)
	} else {
		fmt.Printf("Already seen\n")
	}
}
