package homeassistant

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type HACState byte

const HAC_STATE_INIT HACState = 0
const HAC_STATE_AUTH_SENT HACState = 1
const HAC_STATE_SUBSCRIBE_SENT HACState = 2

type HAWSConsummer struct {
	state HACState
	conn  *websocket.Conn
}

func (wsc *HAWSConsummer) OnConnect(conn *websocket.Conn) error {
	wsc.conn = conn
	return nil
}

func (wsc *HAWSConsummer) OnTextMessage(message []byte) error {
	fmt.Printf("Received: %s\n", message)
	haMessage, err := ParseMessage(message)
	if err != nil {
		return err
	}
	switch wsc.state {
	case HAC_STATE_INIT:
		fmt.Println("Init!")
		if haMessage.GetType() == HA_MSG_TYPE_AUTH_REQUIRED {
			err = wsc.conn.WriteJSON(NewAuthMessage("eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJjNzg5MDNjODJkYmQ0MmM2ODM3MjQ2NzkzZDU3MDkxZCIsImlhdCI6MTU4NjE4MTMxOSwiZXhwIjoxOTAxNTQxMzE5fQ.cw7uXhnefBDkg6luJcgdekUpK2dblI8r5Yv2XomjHzI"))
			if err != nil {
				return err
			}
			wsc.state = HAC_STATE_AUTH_SENT
		}
	case HAC_STATE_AUTH_SENT:
		if haMessage.GetType() == HA_MSG_TYPE_AUTH_INVALID {
			return fmt.Errorf("Auth token invalid")
		}
		if haMessage.GetType() == HA_MSG_TYPE_AUTH_OK {
			err = wsc.conn.WriteJSON(NewSubscribeEventsMessage(1, "deconz_event"))
			if err != nil {
				return err
			}
			wsc.state = HAC_STATE_SUBSCRIBE_SENT
		}
	case HAC_STATE_SUBSCRIBE_SENT:
		result := haMessage.(ResultMessage)
		fmt.Println("Success %s", result.GetSuccess())
	}
	return nil
}

func (wsc *HAWSConsummer) OnBinaryMessage(message []byte) error {
	return nil
}

func (wsc *HAWSConsummer) OnClose() error {
	return nil
}
