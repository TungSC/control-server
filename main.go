package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"control-server/db"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var Redis *db.Redis

const (
	RedisPubSubChannel = "redis_pub_sub_server_resource"
	LiveCdn            = "live-cdn"
)

type UpdateStatusRequest struct {
	Action string       `json:"action"`
	Data   StatusServer `json:"data"`
}

type StatusServer struct {
	Status   string `json:"status"`
	Endpoint string `json:"endpoint"`
}

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
		go healthCheckCdn()

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
					command = LiveCdn
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

func healthCheckCdn() {
	t := time.NewTicker(30 * time.Second)
	for range t.C {
		healthCheckWorker()
	}
}

func healthCheckWorker() {
	output, _ := exec.Command("sh", "-c", fmt.Sprintf("pm2 show %s", LiveCdn)).Output()
	isStopped := strings.Contains(strings.ReplaceAll(string(output), " ", ""), "status│stopped")

	data := UpdateStatusRequest{
		Action: "callback-server",
		Data: StatusServer{
			Status:   "started",
			Endpoint: fmt.Sprintf("https://%s:2443/", os.Getenv("SERVER_ENDPOINT")),
		},
	}

	if isStopped {
		data.Data.Status = "stop"
	}

	requestUpdateStatus(data)
}

func requestUpdateStatus(data UpdateStatusRequest) {
	client := http.Client{}
	reqBody, _ := json.Marshal(data)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%sapi/callback/server", os.Getenv("CMS_ENDPOINT")), bytes.NewBuffer(reqBody))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return
	}

	defer res.Body.Close()
}
