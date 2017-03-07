package main

import (
	"fmt"
	"github.com/minya/e1pm/pmlib"
	"github.com/minya/gopushover"
	"github.com/minya/goutils/config"
	"io/ioutil"
	"os/user"
	"path"
)

func main() {
	var settings Settings
	errSettings := config.UnmarshalJson(&settings, ".e1pm/config.json")
	if nil != errSettings {
		fmt.Printf("Can't read settings: %v\n", errSettings)
		return
	}

	pmClient := pmlib.NewClient()

	setUpErr := pmClient.SetUp(settings.Credentials.Email, settings.Credentials.Password)
	if setUpErr != nil {
		fmt.Printf("Error SetUp: %v\n", setUpErr)
		return
	}

	topicList, topicListErr := pmClient.GetPmTopics()
	if topicListErr != nil {
		fmt.Printf("Error GetTopics: %v\n", topicListErr)
		return
	}

	if len(topicList) == 0 {
		fmt.Println("Error: no topics read\n")
		return
	}

	user, _ := user.Current()
	lastseenPath := path.Join(user.HomeDir, ".e1pm/lastseen.txt")
	dateOfLastSeenPmBin, _ := ioutil.ReadFile(lastseenPath)
	dateOfLastSeenPm := string(dateOfLastSeenPmBin)

	lastTopic := topicList[0]
	dateOfLastPm := lastTopic.Updated
	if dateOfLastPm != dateOfLastSeenPm {
		sendRes, sendErr := gopushover.SendMessage(
			settings.Pushover.Token, settings.Pushover.User, lastTopic.Subject, lastTopic.LastMsg)
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
