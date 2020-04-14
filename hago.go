package hago

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type WSConsumer interface {
	OnConnect(conn *websocket.Conn) error
	OnTextMessage(message []byte) error
	OnBinaryMessage(message []byte) error
	OnClose() error
}

func WSClient(serverUrl *url.URL, consumer WSConsumer) {
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	log.Printf("connecting to %s", serverUrl.String())

	c, _, err := websocket.DefaultDialer.Dial(serverUrl.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})
	if err := consumer.OnConnect(c); err != nil {
		log.Println("consumer onconnect:", err)
		close(interrupt)
	}
	defer func() {
		if err := consumer.OnClose(); err != nil {
			log.Println("consumer onclose:", err)
		}
	}()

	go func() {
		defer close(done)
		for {
			msgType, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			var onMessageErr error = nil
			switch msgType {
			case websocket.BinaryMessage:
				onMessageErr = consumer.OnBinaryMessage(message)
			case websocket.TextMessage:
				onMessageErr = consumer.OnTextMessage(message)
			}
			if onMessageErr != nil {
				log.Println("consumer onmessage:", onMessageErr.Error())
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
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
