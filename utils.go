package main

import (
	"fmt"
	"math/rand"
	"net"
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
