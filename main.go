package main

import (
	"regexp"
	"database/sql"
	"strings"
	"bufio"
	"time"
	"fmt"
	"net"
	"net/textproto"
	"os"
	_ "github.com/go-sql-driver/mysql"
	"github.com/op/go-logging"
)

var (
	db, err = sql.Open("mysql", mysql)
    mainconn net.Conn
	log    = logging.MustGetLogger("example")
	format = logging.MustStringFormatter(
		`%{color}[%{time:2006-01-02 15:04:05}] [%{level:.4s}] %{color:reset}%{message}`,
	)
)

func main() {
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend2 := logging.NewLogBackend(os.Stdout, "", 0)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")
	logging.SetBackend(backend1Leveled, backend2Formatter)

	createConnection()
}

func createConnection() {
    conn, err := net.Dial("tcp", twitchAddress)
    mainconn = conn
	if err != nil {
		log.Error(err)
		reconnect(conn)
		return
	}
	fmt.Fprintf(conn, "PASS %s\r\n", twitchOauth)
	fmt.Fprintf(conn, "USER %s\r\n", twitchUsername)
	fmt.Fprintf(conn, "NICK %s\r\n", twitchUsername)
	 // enable roomstate and such
	fmt.Fprintf(mainconn, "JOIN %s\r\n", "#gempbot")
	log.Debugf("new connection %s", conn.RemoteAddr())
	startDefaultJoin()
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	for {
		line, err := tp.ReadLine()
		if err != nil {
			log.Error(err)
			reconnect(conn)
			break // break loop on errors
		}
		messages := strings.Split(line,"\r\n")
		for _, msg := range messages {
			go parseMessage(msg)
		}

	}
}

func parseMessage(msg string) {
	if !strings.Contains(msg, ".tmi.twitch.tv PRIVMSG #") {
		return
	}
	split1 := strings.Split(msg, ":gempir!gempir@")
	split2 := strings.Split(split1[1], ".tmi.twitch.tv PRIVMSG ")
	username := split2[0]
	rp := regexp.MustCompile(`#\w+\s:`)
	split3 := rp.FindString(split2[1])
	channel := split3[0:len(split3)-2]
	split4 := strings.Split(split2[1], split3)
	message := split4[1]
	rp2 := regexp.MustCompile(`^\x{0001}ACTION\s`)
	rp3 := regexp.MustCompile(`([\x{0001}]+)`)
	message = rp2.ReplaceAllLiteralString(message, "")
	message = rp3.ReplaceAllLiteralString(message, "")
	timestamp := time.Now().Format("2006-01-2 15:04:05")
	saveMessage(channel, username, message, timestamp)
}

func saveMessage(channel, username, message, timestamp string) {
	stmt, err := db.Prepare("INSERT INTO gempLog (channel, username, message, timestamp) VALUES (?, ?, ?, ?)")
    checkErr(err)

    _, err = stmt.Exec(channel, username, message, timestamp)
    checkErr(err)
}

func join(channel string) {
    fmt.Fprintf(mainconn, "JOIN %s\r\n", channel)
}

func startDefaultJoin() {
	rows, err := db.Query("SELECT channel FROM channels")
	checkErr(err)

	for rows.Next() {
		var channel string
		err = rows.Scan(&channel)
		checkErr(err)
		go join(channel)
	}

	defer rows.Close()
}

func checkErr(err error) {
	if err != nil {
		log.Error(err)
	}
}

func reconnect(conn net.Conn) {
	conn.Close()
	createConnection()
}
