package main

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/op/go-logging"
	"net"
	"net/textproto"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	db, err    = sql.Open("mysql", mysql)
	mainconn   net.Conn
	connactive = false
	log        = logging.MustGetLogger("example")
	format     = logging.MustStringFormatter(
		`%{color}[%{time:2006-01-02 15:04:05}] [%{level:.4s}] %{color:reset}%{message}`,
	)
	userrp    = regexp.MustCompile(`:\w+!\w+@\w+\.tmi\.twitch\.tv`)
	channelrp = regexp.MustCompile(`#\w+\s:`)
	actionrp1 = regexp.MustCompile(`^\x{0001}ACTION\s`)
	actionrp2 = regexp.MustCompile(`([\x{0001}]+)`)
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
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("new connection %s", conn.RemoteAddr())
	fmt.Fprintf(conn, "USER %s\r\n", "justinfan123321")
	fmt.Fprintf(conn, "NICK %s\r\n", "justinfan123321")
	// default room
	log.Info("JOIN #gempbot")
	fmt.Fprintf(conn, "JOIN %s\r\n", "#gempbot")
	go startDefaultJoin(conn)

	reader := bufio.NewReader(conn)
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
	defer conn.Close()
}

func parseMessage(msg string) {
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
	timestamp := time.Now().Format("2006-01-2 15:04:05")

	saveMessageToDB(channel, username, message, timestamp)
	saveMessageToTxt(channel, username, message, time.Now())
}

func saveMessageToDB(channel, username, message, timestamp string) {
	_, err := db.Exec("INSERT INTO gempLog (channel, username, message, timestamp) VALUES (?, ?, ?, ?)", channel, username, message, timestamp)
	checkErr(err)
}

func saveMessageToTxt(channel, username, message string, timestamp time.Time) {
	year := timestamp.Year()
	month := timestamp.Month()

	filename := fmt.Sprintf("/var/gemplog/%d/%s/%s.txt", year, month, username)

	log.Debug(filename)

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE,0600)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	contents := fmt.Sprintf("%s[|]%s[|]%s[|]%s\r\n", timestamp.Format("2006-01-2 15:04:05"), channel, username, message)
	log.Debug(contents)
	if _, err = file.WriteString(contents); err != nil {
		log.Error(err)
	}
}


func join(channel string, conn net.Conn) {
	log.Info("JOIN " + channel)
	fmt.Fprintf(conn, "JOIN %s\r\n", channel)
}

func startDefaultJoin(conn net.Conn) {
	rows, err := db.Query("SELECT channel FROM channels")
	checkErr(err)

	for rows.Next() {
		var channel string
		err = rows.Scan(&channel)
		checkErr(err)
		join(channel, conn)
	}

	defer rows.Close()
}

func checkErr(err error) {
	if err != nil {
		log.Error(err)
	}
}
