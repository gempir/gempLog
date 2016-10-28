package main

import (
	"net"
	"os"
	"regexp"

	"github.com/op/go-logging"
	"gopkg.in/redis.v3"
)

var (
	mainconn *net.Conn
	client   *redis.Client
	log      = logging.MustGetLogger("example")
	format   = logging.MustStringFormatter(
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

	conn, err := net.Dial("tcp", twitchAddress)
	if err != nil {
		log.Error(err)
	}

	mainChan := Channel{
		Connection: conn,
		User:       "justinfan123321",
		Nick:       "justinfan123321",
		Room:       "#gempbot",
	}

	mainChan.createConnection()
}

func connectRedis() {
	client = redis.NewClient(&redis.Options{
		Addr:     redisaddress,
		Password: redispass, // no password set
		DB:       0,         // use default DB
	})
	pong, err := client.Ping().Result()
	log.Debug(pong, err)
}

func joinDefault(roomChan chan string) {
	val, err := client.HGetAll("logchannels").Result()
	if err != nil {
		log.Error(err)
	}
	for _, element := range val {
		if element == "1" || element == "0" {
			continue
		}
		roomChan <- element
	}
}

func incUser(username string) {
	client.ZIncrBy("user:lines", 1, username)
}
