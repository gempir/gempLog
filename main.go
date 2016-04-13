package main

import (
	"bufio"
	"fmt"
	"github.com/op/go-logging"
	"net"
	"net/textproto"
	"os"
	"regexp"
	"strings"
	"time"
	"gopkg.in/redis.v3"
)

var (
	mainconn   *net.Conn
	client     *redis.Client
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

	connectRedis()
	createConnection()
}

func connectRedis() {
	client = redis.NewClient(&redis.Options{
        Addr:     redisaddress,
        Password: redispass, // no password set
        DB:       0,  // use default DB
    })
	pong, err := client.Ping().Result()
    log.Debug(pong, err)
}

func createConnection() {
	conn, err := net.Dial("tcp", twitchAddress)
	mainconn = &conn
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

	go joinDefault()

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

func joinDefault() {
	val, err := client.HGetAll("logchannels").Result()
    if err != nil {
        log.Error(err)
    }
    for _, element := range val {
		if element == "1" || element == "0" {
			continue
		}
		go join(element)
	}
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

	saveMessageToTxt(channel, username, message, time.Now())
}


func saveMessageToTxt(channel, username, message string, timestamp time.Time) {
	year := timestamp.Year()
	month := timestamp.Month()

	filename := fmt.Sprintf("/var/gemplog/%d/%s/%s.txt", year, month, username)

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE,0600)
	if err != nil {
		log.Error(err)
	}
	defer file.Close()

	contents := fmt.Sprintf("%s[|]%s[|]%s[|]%s\r\n", timestamp.Format("2006-01-2 15:04:05"), channel, username, message)
	if _, err = file.WriteString(contents); err != nil {
		log.Error(err)
	}
}


func join(channel string) {
	log.Info("JOIN " + channel)
	fmt.Fprintf(*mainconn, "JOIN %s\r\n", channel)
}

func checkErr(err error) {
	if err != nil {
		log.Error(err)
	}
}
