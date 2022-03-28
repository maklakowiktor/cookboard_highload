package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

var orderName = DateNow()
var enc = json.NewEncoder(os.Stdout)
var connected bool = true

const sendToCB = true

func SendMessage(c *websocket.Conn, stringJSON string) {
	err := c.WriteMessage(websocket.TextMessage, []byte(stringJSON))
	if err != nil {
		log.Println("write:", err)
		return
	}
}

func main() {
	fmt.Println("=== Starting Poster simulator ===")

	// Создаем горутину для отслеживания прерывания работы приложения
	ShutdownApp()

	// Сеем зерно для рандома
	rand.Seed(time.Now().UnixNano())

	// Мутим флаги
	flag.Parse()
	log.SetFlags(0)

	// Создаем канал
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Сканируем устройства по ip на порту 2222
	ip, err := ScanDevices(100, 120)
	if err != nil {
		log.Fatal("err dial:", err)
	}

	u := url.URL{Scheme: "ws", Host: ip, Path: "/"}
	log.Printf("connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("err dial:", err)
	}
	defer conn.Close()

	// done := make(chan struct{})

	SendMessage(conn, string(ReadJSONFile("json/handshake.json")))

	InitThread(conn)
}

func InitThread(conn *websocket.Conn) {
	ProcessMessages(conn)

	SendToCookboard(conn)

}

func ReadJSONFile(fileName string) []byte {
	stringJSON, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Successfully opened: %s \n", fileName)

	return stringJSON
}

func ShutdownApp() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		os.Exit(0)
	}()
}

func ProcessMessages(conn *websocket.Conn) {
	go func() {
		// defer close(done)
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

			fmt.Println("Message:", msgMap, "\n")

			if msgMap["action"] == "order_ready" {
				var product Product

				product, err := UnmarshalProduct(message)
				if err != nil {
					panic(err)
				}

				fmt.Println("🍕 Product: ", product, "\n")
			}

			HandleMessage(msgMap, conn, enc)
		}
	}()
}

func HandleMessage(msgMap map[string]interface{}, conn *websocket.Conn, enc *json.Encoder) {
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

		hash := fmt.Sprint(msgMap["hash"])
		terminalId := fmt.Sprint(msgMap["terminalId"])
		receivedMsgHash := fmt.Sprint(msgMap["msgHash"])

		// Answer received
		str := fmt.Sprintf(`{"action": "answer_received", "terminalId": "%s", "isCanceled": 0, "hash": "%s", "msgHash": "%s"}`, terminalId, hash, receivedMsgHash)
		SendMessage(conn, str)

	case "order_cancel":
		fmt.Println("order_canceled:", msgMap)

		hash := fmt.Sprint(msgMap["hash"])
		terminalId := fmt.Sprint(msgMap["terminalId"])
		receivedMsgHash := fmt.Sprint(msgMap["msgHash"])

		str := fmt.Sprintf(`{"action": "answer_received", "terminalId": "%s", "isCanceled": 1, "hash": "%s", "msgHash": "%s"}`, terminalId, hash, receivedMsgHash)
		SendMessage(conn, str)
	default:
		fmt.Println("Произошел flex")
	}
}

func SendToCookboard(conn *websocket.Conn) {
	var interval int32 = 1
	// ticker := time.NewTicker(time.Duration(interval) * time.Second)
	// defer ticker.Stop()

	for {

		if connected && sendToCB {
			orderName++

			const cookedProductsPath = "json/cooked_products.json"
			const allProductsPath = "json/all_products.json"
			const oneProductPath = "json/one_product.json"
			const mixProductsPath = "json/mix_products.json"

			cookedProductsByteArr := ReadJSONFile(mixProductsPath)
			stringJSON := fmt.Sprintf(`
				{
					"id":%d,
					"hash":"%s",
					"type":"workshop",
					"orderName":%d,
					"queueNumber":"A-8",
					"action":"send_order",
					"waiterId":7,
					"waiterName":"Виктор",
					"tableId": 2,
					"account":"web-kotlas",
					"terminalId":
					"web-kotlas1",
					"orderComment":"Комментарий к заказу",
					"products":%s,
					"msgHash":"%s"
				}`, int(DateNow()/1000), RandomString(10), orderName, string(cookedProductsByteArr), RandomString(10))

			SendMessage(conn, stringJSON)
			fmt.Println("Сообщение отправлено ✅")
		}

		interval = rand.Int31n(10-5) + 5
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
