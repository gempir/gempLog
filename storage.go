package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func parseMessage(msg string) {
	log.Debug(msg)

	if !strings.Contains(msg, ".tmi.twitch.tv PRIVMSG ") {
		return
	}

	fulluser := userrp.FindString(msg)
	userirc := strings.Split(fulluser, "!")
	username := userirc[0][1:len(userirc[0])]
	split2 := strings.Split(msg, ".tmi.twitch.tv PRIVMSG ")
	split3 := channelrp.FindString(split2[1])
	channel := split3[0 : len(split3)-2]
	split4 := strings.Split(split2[1], split3)
	message := split4[1]
	message = actionrp1.ReplaceAllLiteralString(message, "")
	message = actionrp2.ReplaceAllLiteralString(message, "")

	log.Infof("[%s] %s @ %s - %s", channel, username, time.Now().Format("2006-01-2 15:04:05"), message)

	incUser(username)
	saveLastMessage(channel, username, message, time.Now())
	saveMessageToTxt(channel, username, message, time.Now())
}

func saveLastMessage(channel, username, message string, timestamp time.Time) {
	channel = strings.Replace(channel, "#", "", 1)
	contents := fmt.Sprintf("%s[|]%s[|]%s[|]%s", timestamp.Format("2006-01-2 15:04:05"), channel, username, message)
	client.HSet("user:lastmessage", username, contents)
}

func saveMessageToTxt(channel, username, message string, timestamp time.Time) {
	year := timestamp.Year()
	month := timestamp.Month()
	channel = strings.Replace(channel, "#", "", 1)
	err := os.MkdirAll(fmt.Sprintf(logfilepath+"%s/%d/%s/", channel, year, month), 0755)
	if err != nil {
		log.Error(err)
		return
	}
	filename := fmt.Sprintf(logfilepath+"%s/%d/%s/%s.txt", channel, year, month, username)

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Error(err)
	}
	defer file.Close()

	contents := fmt.Sprintf("[%s] %s: %s\r\n", timestamp.Format("2006-01-2 15:04:05"), username, message)
	if _, err = file.WriteString(contents); err != nil {
		log.Error(err)
	}
}
