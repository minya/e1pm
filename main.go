package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/user"
	"path"
	"regexp"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(16)

	jar := NewJar()
	transport := DefaultTransport()
	client := http.Client{
		Transport: transport,
		Jar:       jar,
	}

	user, _ := user.Current()
	settingsRoot := path.Join(user.HomeDir, ".e1pm")
	settingsPath := path.Join(settingsRoot, "config.json")
	lastseenPath := path.Join(settingsRoot, "lastseen.txt")

	pmURL := "http://www.e1.ru/talk/forum/pm/"
	pmResult, err := client.Get(pmURL)
	if nil != err {
		fmt.Printf("error fetch %v\n", pmURL)
		return
	}
	pmResult.Body.Close()

	settingsBin, settingsErr := ioutil.ReadFile(settingsPath)
	if settingsErr != nil {
		fmt.Printf("Can't read settings: %v\n", settingsErr)
		return
	}

	var settings Settings
	errSettings := json.Unmarshal(settingsBin, &settings)
	if nil != errSettings {
		fmt.Printf("Can't read settings: %v\n", errSettings)
		return
	}

	loginURL := fmt.Sprintf(
		"https://passport.ngs.ru/e1/login/?redirect_path=%s",
		url.QueryEscape(pmURL))

	values := url.Values{
		"sub":      {"login"},
		"key":      {""},
		"email":    {settings.Credentials.Email},
		"password": {settings.Credentials.Password},
	}

	respLogin, errLogin := client.PostForm(loginURL, values)
	if nil != errLogin {
		fmt.Printf("Error login: %v\n", errLogin)
		return
	}

	tr := transform.NewReader(respLogin.Body, charmap.Windows1251.NewDecoder())

	body, errReadBody := ioutil.ReadAll(tr)
	respLogin.Body.Close()

	if nil != errReadBody {
		fmt.Printf("Can't read body: %v\n", errReadBody)
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
	dateOfLastSeenPmBin, _ := ioutil.ReadFile(lastseenPath)
	if dateOfLastPm != string(dateOfLastSeenPmBin) {
		sendRes, sendErr := SendMessage(settings.Pushover.Token, settings.Pushover.User, "New PM on e1.ru")
		if nil != sendErr {
			fmt.Printf("Can't send push: %v\n", sendErr)
		} else {
			fmt.Printf("Push sent: %v\n", sendRes)
		}
		fmt.Printf("New! %v\n", dateOfLastPm)
		ioutil.WriteFile(lastseenPath, []byte(dateOfLastPm), 0660)
	} else {
		fmt.Printf("Already seen\n")
	}
}
