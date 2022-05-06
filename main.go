package main

import (
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
    "strconv"

    "control-server/db"
    "github.com/joho/godotenv"
)

var Redis *db.Redis

const RedisPubSubChannel = "redis_pub_sub_server_resource"

func main() {
    godotenv.Load()

    dbConfig := os.Getenv("REDIS_DB")
    serverEndpoint := os.Getenv("SERVER_ENDPOINT")
    serverPort := os.Getenv("SERVER_PORT")
    dbNum, _ := strconv.Atoi(dbConfig)

    Redis = db.NewRedis(db.RedisConfig{
        Addrs: []string{os.Getenv("REDIS_SERVER")},
        Pwd:   os.Getenv("REDIS_PASSWORD"),
        DB:    dbNum,
    })

    topic := Redis.Subscribe(RedisPubSubChannel)
    channel := topic.Channel()
    port := fmt.Sprintf("%s:%s", serverEndpoint, serverPort)

    for msg := range channel {
        var receiveMessage db.ReceiveMessage
        // Unmarshal the data into the user
        err := json.Unmarshal([]byte(msg.Payload), &receiveMessage)
        if err != nil {
            _ = fmt.Errorf("json.Unmarshal: %w", err)
        }

        if receiveMessage.Server == port {
            if receiveMessage.Action == "active" {
                exec.Command("sh", "-c", "pm2 start main").Run()
            }

            if receiveMessage.Action == "inactive" {
                exec.Command("sh", "-c", "pm2 stop main").Run()
            }
        }
    }
}
