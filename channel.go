package main

import (
	"bufio"
	"fmt"
	"net"
	"net/textproto"
	"strings"
)

// Channel keeps track of the information specific to a channel connection
type Channel struct {
	Connection net.Conn
	User       string
	Nick       string
	Room       string
}

func (c Channel) createConnection() {
	log.Debugf("new connection %s", c.Connection.RemoteAddr())
	fmt.Fprintf(c.Connection, "USER %s\r\n", c.User)
	fmt.Fprintf(c.Connection, "NICK %s\r\n", c.Nick)
	// default room
	log.Info("JOIN #gempbot")
	fmt.Fprintf(c.Connection, "JOIN %s\r\n", "#gempbot")

	roomChan := make(chan string)

	go joinDefault(roomChan)

	go func() {
		for room := range roomChan {
			c.join(room)
		}
	}()

	reader := bufio.NewReader(c.Connection)
	tp := textproto.NewReader(reader)
	for {
		line, err := tp.ReadLine()
		if err != nil {
			log.Error(err)
			break // break loop on errors
		}
		messages := strings.Split(line, "\r\n")
		if len(messages) == 0 {
			continue
		}
		for _, msg := range messages {
			parseMessage(msg)
		}
	}
	defer c.createConnection() // create new connection when end of conn
}

func (c Channel) join(channel string) {
	log.Info("JOIN " + channel)
	fmt.Fprintf(c.Connection, "JOIN %s\r\n", channel)
	c.Room = channel
}
