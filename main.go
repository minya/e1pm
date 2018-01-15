package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/minya/e1pm/pmlib"
	"github.com/minya/gopushover"
	"github.com/minya/goutils/config"
)

var logPath string

func init() {
	const (
		defaultLogPath = "e1pm.log"
	)
	flag.StringVar(&logPath, "logpath", defaultLogPath, "Path to write logs")

}

func main() {
	flag.Parse()
	SetUpLogger()

	var settings Settings
	errSettings := config.UnmarshalJson(&settings, "~/.e1pm/config.json")
	if nil != errSettings {
		log.Printf("Can't read settings: %v\n", errSettings)
		return
	}

	pmClient := pmlib.NewClient()

	setUpErr := pmClient.SetUp(settings.Credentials.Email, settings.Credentials.Password)
	if setUpErr != nil {
		log.Printf("Error SetUp: %v\n", setUpErr)
		return
	}

	topicList, topicListErr := pmClient.GetPmTopics()
	if topicListErr != nil {
		log.Printf("Error GetTopics: %v\n", topicListErr)
		return
	}

	if len(topicList) == 0 {
		log.Println("Error: no topics read\n")
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
			log.Printf("Can't send push: %v\n", sendErr)
		} else {
			log.Printf("Push sent: %v\n", sendRes)
		}
		log.Printf("New! %v\n", dateOfLastPm)
		ioutil.WriteFile(lastseenPath, []byte(dateOfLastPm), 0660)
	} else {
		log.Printf("Already seen\n")
	}
}

func SetUpLogger() {
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(logFile)
}
