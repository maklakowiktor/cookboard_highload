package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", fmt.Sprintf("%s:2222", GetLocalIP()), "http service address")

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	conn, err := net.Dial("ip:icmp", "google.com")
	if err != nil {
		return ""
	}
	var localIp = conn.LocalAddr()
	// fmt.Println(localIp)
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

func SendMessage(c *websocket.Conn, stringJSON string) {
	err := c.WriteMessage(websocket.TextMessage, []byte(stringJSON))
	if err != nil {
		log.Println("write:", err)
		return
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	log.SetFlags(0)
	connected := false

	fmt.Println("Local ip: ", GetLocalIP())

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Printf("connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	done := make(chan struct{})

	stringJSON := `{"action":"handshake","accountName":"web-kotlas","terminalId":"стресс-тест","type":"FASTFOOD","msgHash":"WddYGbBgAy"}`
	SendMessage(conn, stringJSON)

	orderName := 2000

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			if !connected {
				connected = true
			}

			log.Printf("recv: %s", message)

		}
	}()

	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:

			if connected {
				orderName++

				stringJSON = fmt.Sprintf(`{"id":%d,"hash":"%s","type":"workshop","orderName":%d,"action":"send_order","waiterId":7,"waiterName":"Виктор","tableId":"","account":"web-kotlas","terminalId":"web-kotlas1","comment":"KITCHEN Bar","orderComment":"stress","products":[{"id":9,"feedPriority":1,"count":1,"modification":"[{\"m\":1,\"a\":2},{\"m\":3,\"a\":1}]","name":"Большой Денер (белый соус)","cookingTime":200,"title":"Мясо × 2, Огурцы маринованные","titleArray":["Мясо × 2","Огурцы маринованные"],"productId":"16431463320039[{\"m\":1,\"a\":2},{\"m\":3,\"a\":1}]1","comment":"KITCHEN Bar"},{"id":5,"feedPriority":2,"count":8,"name":"Круасаны","cookingTime":150,"title":"","titleArray":[],"productId":"164314633200352","comment":"KITCHEN Bar"},{"id":5,"feedPriority":3,"count":4,"name":"Круасаны","cookingTime":150,"title":"","titleArray":[],"productId":"164314633200353","comment":"KITCHEN Bar"},{"id":5,"feedPriority":1,"count":3,"name":"Круасаны","cookingTime":150,"title":"","titleArray":[],"productId":"164314633200351","comment":"KITCHEN Bar"},{"id":3,"feedPriority":1,"count":4,"name":"Капучино 250 мл","cookingTime":80,"title":"","titleArray":[],"productId":"164314633200331","comment":"KITCHEN Bar"},{"id":3,"feedPriority":2,"count":7,"name":"Капучино 250 мл","cookingTime":80,"title":"","titleArray":[],"productId":"164314633200332","comment":"KITCHEN Bar"},{"id":3,"feedPriority":3,"count":3,"name":"Капучино 250 мл","cookingTime":80,"title":"","titleArray":[],"productId":"%s","comment":"KITCHEN Bar"}],"msgHash":"%s"}`, DateNow(), RandomString(10), orderName, RandomString(10), RandomString(10))

				// stringJSON := fmt.Sprintf(`{"id": %d, "hash": "%s", "type": "workshop", "orderName": %d, "action": "send_order", "waiterId": 7, "waiterName": "Виктор", "tableId": "99", "account": "web-kotlas", "terminalId": "web-kotlas1", "comment": "стресс коммент", "orderComment": "", "products": [{"id": 3, "count": 1, "name": "Капучино 250 мл", "cookingTime": 80, "title": "", "titleArray": [], "productId": "%s", "comment": ""}], "msgHash": "%s"}`, DateNow(), RandomString(10), orderName, RandomString(10), RandomString(10))

				SendMessage(conn, stringJSON)
			}

			// {"action":"handshake","accountName":"web-kotlas","terminalId":"web-kotlas1","type":"FASTFOOD","msgHash":"WddYGbBgAy"}
			// log.Println("time: ", t)

		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
