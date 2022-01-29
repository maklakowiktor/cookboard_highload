package main

import (
	"encoding/json"
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

// 192.168.1.100, 192.168.1.110
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

func HandleMessage(msgMap map[string]interface{}, conn *websocket.Conn, enc *json.Encoder) {
	// fmt.Print(msgMap["action"], ": ")
	switch msgMap["action"] {
	case "handshake":
	case "transportMsgReceived":
		// fmt.Print(msgMap)
		// 1. {action: transportMsgReceived, receivedMsgHash: 2Aad9Y3oanIB, msgHash: nym9Ualy0N}

		// receivedMsgHash := fmt.Sprint(msgMap["receivedMsgHash"])
		// fmt.Println("receivedMsgHash: ", receivedMsgHash)
		// str := fmt.Sprintf(`{"action": "transportMsgReceived", "receivedMsgHash": "%s", "msgHash": "%s"}`, receivedMsgHash, RandomString(12))
		// SendMessage(conn, str)

	case "order_ready":
		// 2. {action: answer_received, terminalId: web-kotlas1, isCanceled: 0, hash: mTL5tyJJFaed, msgHash: LlRc9OGEOd}
		msgJson := fmt.Sprint(msgMap["msg"])
		// fmt.Println("order_ready:", msgJson)

		var msg map[string]interface{}

		if err := json.Unmarshal([]byte(msgJson), &msg); err != nil {
			panic(err)
		}
		// fmt.Println()
		// fmt.Println(msg)

		hash := fmt.Sprint(msg["hash"])
		terminalId := fmt.Sprint(msgMap["terminalId"])
		receivedMsgHash := fmt.Sprint(msgMap["msgHash"])

		// Transport
		// str := fmt.Sprintf(`{"action": "transportMsgReceived", "receivedMsgHash": "%s", "msgHash": "%s"}`, receivedMsgHash, RandomString(12))
		// SendMessage(conn, str)

		// Answer received
		str := fmt.Sprintf(`{"action": "answer_received", "terminalId": "%s", "isCanceled": 0, "hash": "%s", "msgHash": "%s"}`, terminalId, hash, receivedMsgHash)
		SendMessage(conn, str)

	case "order_canceled":
		msgJson := fmt.Sprint(msgMap["msg"])
		fmt.Println("order_canceled:", msgJson)

		var msg map[string]interface{}

		if err := json.Unmarshal([]byte(msgJson), &msg); err != nil {
			panic(err)
		}

		hash := fmt.Sprint(msgMap["hash"])
		terminalId := fmt.Sprint(msgMap["terminalId"])
		receivedMsgHash := fmt.Sprint(msgMap["receivedMsgHash"])

		str := fmt.Sprintf(`{"action": "answer_received", "terminalId": "%s", "isCanceled": 1, "hash": "%s", "msgHash": "%s"}`, terminalId, hash, receivedMsgHash)
		SendMessage(conn, str)
	default:
		fmt.Println("Произошел flex")
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
	enc := json.NewEncoder(os.Stdout)

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

			// log.Printf("recv: %s", message)

			var msgMap map[string]interface{}

			if err := json.Unmarshal(message, &msgMap); err != nil {
				panic(err)
			}

			fmt.Println("Message:", msgMap)

			HandleMessage(msgMap, conn, enc)

			// fmt.Println(message["action"])

			// d := map[string]int{"apple": 5, "lettuce": 7}
			// enc.Encode(d)

		}
	}()

	var interval int32 = 1
	// ticker := time.NewTicker(time.Duration(interval) * time.Second)
	// defer ticker.Stop()

	for {
		if connected {
			orderName++

			// stringJSON = fmt.Sprintf(`{"id":%d,"hash":"%s","type":"workshop","orderName":%d,"action":"send_order","waiterId":7,"waiterName":"Виктор","tableId":"","account":"web-kotlas","terminalId":"web-kotlas1","comment":"KITCHEN Bar","orderComment":"stress","products":[{"id":9,"feedPriority":1,"count":1,"modification":"[{\"m\":1,\"a\":2},{\"m\":3,\"a\":1}]","name":"Большой Денер (белый соус)","cookingTime":200,"title":"Мясо × 2, Огурцы маринованные","titleArray":["Мясо × 2","Огурцы маринованные"],"productId":"16431463320039[{\"m\":1,\"a\":2},{\"m\":3,\"a\":1}]1","comment":"KITCHEN Bar"}],"msgHash":"%s"}`, DateNow(), RandomString(10), orderName, RandomString(10))
			stringJSON = fmt.Sprintf(`{"id":%d,"hash":"%s","type":"workshop","orderName":%d,"action":"send_order","waiterId":7,"waiterName":"Виктор","tableId":"","account":"web-kotlas","terminalId":"web-kotlas1","comment":"KITCHEN Bar","orderComment":"stress","products":[{"id":9,"feedPriority":1,"count":1,"modification":"[{\"m\":1,\"a\":2},{\"m\":3,\"a\":1}]","name":"Большой Денер (белый соус)","cookingTime":200,"title":"Мясо × 2, Огурцы маринованные","titleArray":["Мясо × 2","Огурцы маринованные"],"productId":"16431463320039[{\"m\":1,\"a\":2},{\"m\":3,\"a\":1}]1","comment":"KITCHEN Bar"},{"id":5,"feedPriority":2,"count":8,"name":"Круасаны","cookingTime":150,"title":"","titleArray":[],"productId":"164314633200352","comment":"KITCHEN Bar"},{"id":5,"feedPriority":3,"count":4,"name":"Круасаны","cookingTime":150,"title":"","titleArray":[],"productId":"164314633200353","comment":"KITCHEN Bar"},{"id":5,"feedPriority":1,"count":3,"name":"Круасаны","cookingTime":150,"title":"","titleArray":[],"productId":"164314633200351","comment":"KITCHEN Bar"},{"id":3,"feedPriority":1,"count":4,"name":"Капучино 250 мл","cookingTime":80,"title":"","titleArray":[],"productId":"164314633200331","comment":"KITCHEN Bar"},{"id":3,"feedPriority":2,"count":7,"name":"Капучино 250 мл","cookingTime":80,"title":"","titleArray":[],"productId":"164314633200332","comment":"KITCHEN Bar"},{"id":3,"feedPriority":3,"count":3,"name":"Капучино 250 мл","cookingTime":80,"title":"","titleArray":[],"productId":"%s","comment":"KITCHEN Bar"}],"msgHash":"%s"}`, DateNow(), RandomString(10), orderName, RandomString(10), RandomString(10))

			// stringJSON := fmt.Sprintf(`{"id": %d, "hash": "%s", "type": "workshop", "orderName": %d, "action": "send_order", "waiterId": 7, "waiterName": "Виктор", "tableId": "99", "account": "web-kotlas", "terminalId": "web-kotlas1", "comment": "стресс коммент", "orderComment": "", "products": [{"id": 3, "count": 1, "name": "Капучино 250 мл", "cookingTime": 80, "title": "", "titleArray": [], "productId": "%s", "comment": ""}], "msgHash": "%s"}`, DateNow(), RandomString(10), orderName, RandomString(10), RandomString(10))

			// SendMessage(conn, stringJSON)
		}
		fmt.Println("Interval: ", interval)
		time.Sleep(time.Duration(interval) * time.Second)
		interval = rand.Int31n(15-5) + 5

		// select {
		// case <-done:
		// 	return
		// case <-ticker.C:

		// 	if connected {
		// 		orderName++

		// 		stringJSON = fmt.Sprintf(`{"id":%d,"hash":"%s","type":"workshop","orderName":%d,"action":"send_order","waiterId":7,"waiterName":"Виктор","tableId":"","account":"web-kotlas","terminalId":"web-kotlas1","comment":"KITCHEN Bar","orderComment":"stress","products":[{"id":9,"feedPriority":1,"count":1,"modification":"[{\"m\":1,\"a\":2},{\"m\":3,\"a\":1}]","name":"Большой Денер (белый соус)","cookingTime":200,"title":"Мясо × 2, Огурцы маринованные","titleArray":["Мясо × 2","Огурцы маринованные"],"productId":"16431463320039[{\"m\":1,\"a\":2},{\"m\":3,\"a\":1}]1","comment":"KITCHEN Bar"},{"id":5,"feedPriority":2,"count":8,"name":"Круасаны","cookingTime":150,"title":"","titleArray":[],"productId":"164314633200352","comment":"KITCHEN Bar"},{"id":5,"feedPriority":3,"count":4,"name":"Круасаны","cookingTime":150,"title":"","titleArray":[],"productId":"164314633200353","comment":"KITCHEN Bar"},{"id":5,"feedPriority":1,"count":3,"name":"Круасаны","cookingTime":150,"title":"","titleArray":[],"productId":"164314633200351","comment":"KITCHEN Bar"},{"id":3,"feedPriority":1,"count":4,"name":"Капучино 250 мл","cookingTime":80,"title":"","titleArray":[],"productId":"164314633200331","comment":"KITCHEN Bar"},{"id":3,"feedPriority":2,"count":7,"name":"Капучино 250 мл","cookingTime":80,"title":"","titleArray":[],"productId":"164314633200332","comment":"KITCHEN Bar"},{"id":3,"feedPriority":3,"count":3,"name":"Капучино 250 мл","cookingTime":80,"title":"","titleArray":[],"productId":"%s","comment":"KITCHEN Bar"}],"msgHash":"%s"}`, DateNow(), RandomString(10), orderName, RandomString(10), RandomString(10))

		// 		// stringJSON := fmt.Sprintf(`{"id": %d, "hash": "%s", "type": "workshop", "orderName": %d, "action": "send_order", "waiterId": 7, "waiterName": "Виктор", "tableId": "99", "account": "web-kotlas", "terminalId": "web-kotlas1", "comment": "стресс коммент", "orderComment": "", "products": [{"id": 3, "count": 1, "name": "Капучино 250 мл", "cookingTime": 80, "title": "", "titleArray": [], "productId": "%s", "comment": ""}], "msgHash": "%s"}`, DateNow(), RandomString(10), orderName, RandomString(10), RandomString(10))

		// 		SendMessage(conn, stringJSON)
		// 	}

		// 	interval = rand.Int31n(120-5) + 5
		// 	fmt.Println("Interval: ", interval)
		// 	// {"action":"handshake","accountName":"web-kotlas","terminalId":"web-kotlas1","type":"FASTFOOD","msgHash":"WddYGbBgAy"}
		// 	// log.Println("time: ", t)

		// case <-interrupt:
		// 	log.Println("interrupt")

		// 	// Cleanly close the connection by sending a close message and then
		// 	// waiting (with timeout) for the server to close the connection.
		// 	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		// 	if err != nil {
		// 		log.Println("write close:", err)
		// 		return
		// 	}
		// 	select {
		// 	case <-done:
		// 	case <-time.After(time.Second):
		// 	}
		// 	return
		// }
	}
}
