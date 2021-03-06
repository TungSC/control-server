package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strconv"

	"control-server/db"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var Redis *db.Redis

const RedisPubSubChannel = "redis_pub_sub_server_resource"

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Accept", "Accept-Language", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	dbConfig := os.Getenv("REDIS_DB")
	serverEndpoint := os.Getenv("SERVER_ENDPOINT")
	dbNum, _ := strconv.Atoi(dbConfig)

	go func() {
		Redis = db.NewRedis(db.RedisConfig{
			Addrs: []string{os.Getenv("REDIS_SERVER")},
			Pwd:   os.Getenv("REDIS_PASSWORD"),
			DB:    dbNum,
		})

		go db.UsageData(Redis)

		topic := Redis.Subscribe(RedisPubSubChannel)
		channel := topic.Channel()

		for msg := range channel {
			var receiveMessage db.ReceiveMessage
			// Unmarshal the data into the user
			err := json.Unmarshal([]byte(msg.Payload), &receiveMessage)
			if err != nil {
				_ = fmt.Errorf("json.Unmarshal: %w", err)
			}

			// service
			if receiveMessage.Server == "" {
				continue
			}

			newUrl := url.URL{
				Host: receiveMessage.Server,
			}

			if newUrl.Hostname() == serverEndpoint {
				command := ""
				port := newUrl.Port()

				switch port {
				case "1935":
					command = "live-srs"
				case "1954":
					command = "cdn-main"
				case "2443":
					command = "cdn-main"
				default:
					command = "pegatv-transcode-dev-live-1"
				}

				// service's child
				if receiveMessage.Action == "active" {
					exec.Command("sh", "-c", fmt.Sprintf("pm2 start %s", command)).Run()
				}

				if receiveMessage.Action == "inactive" {
					exec.Command("sh", "-c", fmt.Sprintf("pm2 stop %s", command)).Run()
				}
			}
		}
	}()

	log.Info("Server is running...")
	err := r.Run()
	if err != nil {
		log.Fatalf("Starting server: %s", err)
	}
}
