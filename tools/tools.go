package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func Http_GET() {
	url := "http://utools.run/hardware_inventory_win7"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("GET call failed")
	}
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("client: response body: %s\n", resBody)
}

type Message struct {
	Message string `json:"Message"`
}

func Http_POST(msg string) {
	requestURL := "http://utools.run/mylog"

	message := Message{
		Message: msg,
	}
	jsonBody, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
	}

	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)
	if err != nil {
		fmt.Println("http POST error")
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("read response data err")
	}
	fmt.Println("post response is:", string(data))
}
