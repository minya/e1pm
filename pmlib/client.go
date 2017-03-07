package pmlib

import (
	"fmt"
	"github.com/minya/goutils/web"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/http"
	"net/url"
)

type PmClient struct {
	httpClient *http.Client
	pmHomeUrl  string
}

func NewClient() *PmClient {
	pmClient := &PmClient{}
	jar := web.NewJar()
	transport := web.DefaultTransport(1000)

	pmClient.httpClient = &http.Client{
		Transport: transport,
		Jar:       jar,
	}

	pmClient.pmHomeUrl = "http://www.e1.ru/talk/forum/pm/"
	return pmClient
}

func (self *PmClient) SetUp(login string, password string) error {
	return self.setUp(login, password)
}

func (self *PmClient) GetPmTopics() ([]PmTopic, error) {
	return self.getPmTopics()
}

func (self *PmClient) setUp(login string, password string) error {
	_, _ = self.httpClient.Get(self.pmHomeUrl)

	loginData := url.Values{
		"sub":      {"login"},
		"key":      {""},
		"url":      {self.pmHomeUrl},
		"login":    {login},
		"password": {password},
	}

	loginQuery := url.Values{
		"token_name": {"ngs_token"},
		"return":     {self.pmHomeUrl},
	}

	loginUrl := fmt.Sprintf("https://passport.ngs.ru/login/?%v", loginQuery.Encode())

	_, errLogin := self.httpClient.PostForm(loginUrl, loginData)
	if nil != errLogin {
		fmt.Printf("Error login: %v\n", errLogin)
		return nil
	}
	return nil
}

func (self *PmClient) getPmTopics() ([]PmTopic, error) {
	respPm, errPm := self.httpClient.Get(self.pmHomeUrl)
	if nil != errPm {
		return nil, errPm
	}

	tr := transform.NewReader(respPm.Body, charmap.Windows1251.NewDecoder())

	body, errReadBody := ioutil.ReadAll(tr)
	respPm.Body.Close()

	if nil != errReadBody {
		return nil, errReadBody
	}

	html := string(body)
	topicList := ParseTopicsList(html)
	return topicList, nil
}
