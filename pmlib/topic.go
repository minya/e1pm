package pmlib

import (
	"regexp"
)

func ParseTopicsList(html string) []PmTopic {
	regex, _ := regexp.Compile(
		//"<span class=\"text_orange\"><strong>Новое!</strong></span><br><span class=\"small_gray\">(.*)</span>")
		"<a href=\"/talk/forum/pm/d/.+?/\">(.+?)</a>[.\\S\\s]+?<strong><.+?>(.+?)</.+?>[.\\s\\S]+?<span class=\"small_gray\">(.+?)</span><br>(.+?)</td>")
	match := regex.FindAllStringSubmatch(html, -1)
	if len(match) < 1 {
		//fmt.Println("No new messages matched")
		return nil
	}

	result := make([]PmTopic, len(match))

	for idx, item := range match {
		topic := PmTopic{
			Subject: item[1],
			Who:     item[2],
			Updated: item[3],
			LastMsg: item[4],
		}
		result[idx] = topic
	}

	return result
}

type PmTopic struct {
	Subject string
	Who     string
	Updated string
	LastMsg string
}
