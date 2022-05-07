package db

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const RedisPrefixServerUsage = "ovp_nodejs:Server:resource:"

type result struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Result []float64 `json:"result"`
}

type usageDataResponse struct {
	Mem float64 `json:"mem"`
	CPU float64 `json:"cpu"`
	Net struct {
		Inbound  float64 `json:"inbound"`
		Outbound float64 `json:"outbound"`
	} `json:"net"`
}

func UsageData(redis *Redis) {
	t := time.NewTicker(1 * time.Second)
	for range t.C {
		worker(redis)
	}
}

func worker(redis *Redis) {
	var usedData usageDataResponse
	var serverEndpoint = os.Getenv("SERVER_ENDPOINT")
	var serverPort = os.Getenv("SERVER_PORT")

	ram, err := getRamUsage()
	if err != nil {
		usedData.Mem = 0
	} else {
		usedData.Mem = ram.Result[0]
	}

	cpu, err := getCpuUsage()
	if err != nil {
		usedData.CPU = 0
	} else {
		usedData.CPU = cpu.Result[0]
	}

	inBound, err := getInBound()
	if err != nil {
		usedData.Net.Inbound = 0
	} else {
		usedData.Net.Inbound = inBound.Result[0]
	}

	outBound, err := getOutBound()
	if err != nil {
		usedData.Net.Outbound = 0
	} else {
		usedData.Net.Outbound = outBound.Result[0]
	}

	jsonData, _ := json.Marshal(usedData)
	key := fmt.Sprintf("%s%s:%s", RedisPrefixServerUsage, serverEndpoint, serverPort)
	err = redis.Del(key)
	if err != nil {
		return
	}

	err = redis.Set(key, string(jsonData), -1)
	if err != nil {
		return
	}
}

func getData(endpoint string) (*result, error) {
	client := http.Client{
		Timeout: time.Second * 3, // Timeout after 3 seconds
	}
	req, _ := http.NewRequest(http.MethodGet, endpoint, nil)

	req.Header.Set("Connection", "keep-alive")
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("json.NewDecoder: %w", err)
	}
	defer res.Body.Close()

	var response *result
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("json.NewDecoder: %w", err)
	}

	return response, nil
}

func getRamUsage() (*result, error) {
	url := "http://127.0.0.1:19999/api/v1/data?chart=system.ram&format=array&points=1&group=average&gtime=0&options=absolute|percentage|jsonwrap|nonzero&after=-1&dimensions=used|buffers|active|wired"
	return getData(url)
}

func getCpuUsage() (*result, error) {
	url := "http://127.0.0.1:19999/api/v1/data?chart=system.cpu&format=array&points=1&group=average&gtime=0&options=absolute|jsonwrap|nonzero&after=-1"
	return getData(url)
}

func getInBound() (*result, error) {
	url := "http://127.0.0.1:19999/api/v1/data?chart=system.net&format=array&points=1&group=average&gtime=0&options=absolute|jsonwrap|nonzero&after=-1&dimensions=received"
	return getData(url)
}

func getOutBound() (*result, error) {
	url := "http://127.0.0.1:19999/api/v1/data?chart=system.net&format=array&points=1&group=average&gtime=0&options=absolute|jsonwrap|nonzero&after=-1&dimensions=sent"
	return getData(url)
}
