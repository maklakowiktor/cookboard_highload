package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"time"
)

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	conn, err := net.Dial("ip:icmp", "google.com")
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	var localIp = conn.LocalAddr()
	fmt.Println("LocIP: ", localIp)
	return localIp.String()
}

func DateNow() int64 {
	now := time.Now()

	// correct way to convert time to millisecond - with UnixNano()
	unixNano := now.UnixNano()
	umillisec := unixNano / 1000000
	// fmt.Println("(correct)Millisecond : ", umillisec)
	return umillisec
}

func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func ScanDevices(startOctet, endOctet int) (string, error) {
	const PORT = 2222

	for i := startOctet; i < endOctet; i++ {
		var ip string = fmt.Sprintf("192.168.1.%s", fmt.Sprint(i))
		var addr = fmt.Sprintf("%s:%s", ip, fmt.Sprint(PORT))
		// fmt.Println("Current addr: ", addr)

		client := &http.Client{Timeout: 300 * time.Millisecond}

		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s", addr), nil)
		if err != nil {
			continue
		}

		req.Header.Add("Accept", "application/json")
		resp, err := client.Do(req)

		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		fmt.Println("addr: ", addr, ", response: ", string(body))

		if resp.StatusCode == 200 {
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			if result["ip"] == ip {
				return addr, nil
			}
		}

		defer client.CloseIdleConnections()

	}

	return "", errors.New("device not found")
}
