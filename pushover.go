package main

import (
	//"errors"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// PushoverResult struct represents response from pushover
type PushoverResult struct {
	Status  int    `json:"status"`
	Request string `json:"request"`
}

// SendMessage sends message to pushover
func SendMessage(token string, user string, msg string) (res *PushoverResult, e error) {
	transport := DefaultTransport()
	client := http.Client{
		Transport: transport,
	}

	form := url.Values{
		"message": {msg},
		"token":   {token},
		"user":    {user},
	}

	resp, err := client.PostForm("https://api.pushover.net/1/messages.json", form)
	if err != nil {
		return nil, err
	}

	fmt.Printf("resp from pushover: %v\n", resp.StatusCode)

	bodyBin, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := new(PushoverResult)
	err = json.Unmarshal(bodyBin, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
