package hago

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/enolgor/hago/homeassistant"
	"github.com/gorilla/websocket"
)

type clientState byte

const stateInit clientState = 0
const stateAuthSent clientState = 1
const stateSubscribeSent clientState = 2
const stateListen clientState = 3

type clientImp struct {
	state              clientState
	conn               *websocket.Conn
	url                *url.URL
	authToken          string
	eventListeners     map[ListenerHandler]func(*homeassistant.Event)
	eventSubscriptions map[string][]ListenerHandler
	eventListenerCount uint16
	stateListeners     map[ListenerHandler]func(*homeassistant.StateChangedData)
	stateSubscriptions map[string][]ListenerHandler
	stateListenerCount uint16
	onConnect          func()
	closeComms         chan struct{}
	commandCount       int
	commandClosers     map[int]chan *homeassistant.ResultMessage
	states             map[string]*homeassistant.State
}

func NewClient(url *url.URL, authToken string) Client {
	c := &clientImp{
		state:              stateInit,
		conn:               nil,
		url:                url,
		authToken:          authToken,
		eventListeners:     make(map[ListenerHandler]func(*homeassistant.Event)),
		eventSubscriptions: make(map[string][]ListenerHandler),
		eventListenerCount: 0,
		stateListeners:     make(map[ListenerHandler]func(*homeassistant.StateChangedData)),
		stateSubscriptions: make(map[string][]ListenerHandler),
		stateListenerCount: 0,
		closeComms:         make(chan struct{}),
		commandCount:       1, // starts with 1 because first command is subscribe to all (see below)
		commandClosers:     make(map[int]chan *homeassistant.ResultMessage),
	}
	return c
}

func (c *clientImp) Connect(onConnect func(), onClose func()) {
	c.onConnect = onConnect
	go wsClient(c.url, func(conn *websocket.Conn) {
		c.conn = conn
	}, onClose, func(rawMessage []byte) {
		message, err := homeassistant.ParseMessage(rawMessage)
		if err != nil {
			log.Println("error parsing message", err)
		}
		err = c.handleMessage(message)
		if err != nil {
			log.Println("error handling message", err)
		}
	}, c.closeComms)
}

func (c *clientImp) SubscribeToEvent(eventType string, onEvent func(*homeassistant.Event)) ListenerHandler {
	c.eventListenerCount++
	handler := ListenerHandler(c.eventListenerCount)
	c.eventListeners[handler] = onEvent
	listenerHandlers, ok := c.eventSubscriptions[eventType]
	if !ok {
		listenerHandlers = []ListenerHandler{}
	}
	listenerHandlers = append(listenerHandlers, handler)
	c.eventSubscriptions[eventType] = listenerHandlers
	return handler
}

func (c *clientImp) UnsubscribeFromEvent(handler ListenerHandler) {
	_, ok := c.eventListeners[handler]
	if ok {
		delete(c.eventListeners, handler)
	}
}

func (c *clientImp) dispatchEvent(event *homeassistant.Event) {
	if listenerHandlers, ok := c.eventSubscriptions[event.EventType]; ok {
		var f func(*homeassistant.Event)
		for _, handler := range listenerHandlers {
			if f, ok = c.eventListeners[handler]; ok {
				go f(event)
			}
		}
	}
}

func (c *clientImp) SubscribeToState(entityID string, onStateChange func(*homeassistant.StateChangedData)) ListenerHandler {
	c.stateListenerCount++
	handler := ListenerHandler(c.stateListenerCount)
	c.stateListeners[handler] = onStateChange
	listenerHandlers, ok := c.eventSubscriptions[entityID]
	if !ok {
		listenerHandlers = []ListenerHandler{}
	}
	listenerHandlers = append(listenerHandlers, handler)
	c.stateSubscriptions[entityID] = listenerHandlers
	return handler
}

func (c *clientImp) UnsubscribeFromState(handler ListenerHandler) {
	_, ok := c.stateListeners[handler]
	if ok {
		delete(c.stateListeners, handler)
	}
}

func (c *clientImp) dispatchStateChange(stateChangedData *homeassistant.StateChangedData) {
	if listenerHandlers, ok := c.stateSubscriptions[stateChangedData.EntityID]; ok {
		var f func(*homeassistant.StateChangedData)
		for _, handler := range listenerHandlers {
			if f, ok = c.stateListeners[handler]; ok {
				go f(stateChangedData)
			}
		}
	}
}

func (c *clientImp) CallService(domain string, service string, data map[string]interface{}) error {
	callServiceMessage := homeassistant.NewMessage(homeassistant.HA_MSG_TYPE_CALL_SERVICE).(*homeassistant.CallServiceMessage)
	callServiceMessage.Domain = domain
	callServiceMessage.Service = service
	callServiceMessage.Data = data
	resultMessage, err := c.sendCommand(callServiceMessage)
	if err != nil {
		return err
	}
	if !resultMessage.Success {
		return resultMessage.Error
	}
	return nil
}

func (c *clientImp) FetchStates() (map[string]homeassistant.State, error) {
	getStatesMessage := homeassistant.NewMessage(homeassistant.HA_MSG_TYPE_GET_STATES).(*homeassistant.GetStatesMessage)
	resultMessage, err := c.sendCommand(getStatesMessage)
	if err != nil {
		return nil, err
	}
	if !resultMessage.Success {
		return nil, resultMessage.Error
	}
	stateMap := make(map[string]homeassistant.State)
	for _, state := range resultMessage.Result.States {
		stateMap[state.EntityID] = state
	}
	return stateMap, nil
}

func (c *clientImp) sendCommand(message homeassistant.Message) (*homeassistant.ResultMessage, error) {
	c.commandCount++
	closer := make(chan *homeassistant.ResultMessage)
	c.commandClosers[c.commandCount] = closer
	message.SetID(c.commandCount)
	err := c.sendMessage(message)
	if err != nil {
		return nil, err
	}
	resultMessage := <-closer
	close(closer)
	return resultMessage, nil
}

func (c *clientImp) handleResultMessage(resultMessage *homeassistant.ResultMessage) {
	if closer, ok := c.commandClosers[resultMessage.GetID()]; ok {
		closer <- resultMessage
		delete(c.commandClosers, resultMessage.GetID())
	}
}

func (c *clientImp) Close() {
	close(c.closeComms)
}

func (c *clientImp) sendMessage(message homeassistant.Message) error {
	err := c.conn.WriteJSON(message)
	return err
}

func (c *clientImp) handleMessage(message homeassistant.Message) error {
	switch c.state {
	case stateInit:
		if message.GetType() == homeassistant.HA_MSG_TYPE_AUTH_REQUIRED {
			authMessage := homeassistant.NewMessage(homeassistant.HA_MSG_TYPE_AUTH).(*homeassistant.AuthMessage)
			authMessage.AccessToken = c.authToken
			err := c.sendMessage(authMessage)
			if err != nil {
				return err
			}
			c.state = stateAuthSent
		}
	case stateAuthSent:
		if message.GetType() == homeassistant.HA_MSG_TYPE_AUTH_INVALID {
			defer close(c.closeComms)
			return fmt.Errorf("Auth token invalid")
		}
		if message.GetType() == homeassistant.HA_MSG_TYPE_AUTH_OK {
			subscribeEventsMessage := homeassistant.NewMessage(homeassistant.HA_MSG_TYPE_SUBSCRIBE_EVENTS).(*homeassistant.SubscribeEventsMessage)
			subscribeEventsMessage.SetID(1) // first command is subscribe to all
			subscribeEventsMessage.EventType = nil
			err := c.sendMessage(subscribeEventsMessage)
			if err != nil {
				return err
			}
			c.state = stateSubscribeSent
		}
	case stateSubscribeSent:
		if message.GetType() == homeassistant.HA_MSG_TYPE_RESULT {
			result := message.(*homeassistant.ResultMessage)
			if result.GetID() == 1 {
				if result.Success {
					go c.onConnect()
					c.state = stateListen
					return nil
				}
				return fmt.Errorf("Unsuccess subscribe")
			}
		}
	case stateListen:
		if message.GetType() == homeassistant.HA_MSG_TYPE_EVENT {
			eventMessage := message.(*homeassistant.EventMessage)
			if eventMessage.GetID() == 1 {
				if eventMessage.Event.EventType == "state_changed" {
					c.dispatchStateChange(eventMessage.Event.StateChangedData)
				} else {
					c.dispatchEvent(eventMessage.Event)
				}
			} else {
				return fmt.Errorf("Unknown event id channel: %d", eventMessage.ID)
			}
			return nil
		}
		if message.GetType() == homeassistant.HA_MSG_TYPE_RESULT {
			resultMessage := message.(*homeassistant.ResultMessage)
			c.handleResultMessage(resultMessage)
			return nil
		}
		return fmt.Errorf("Unexpect message type %s", message.GetType())
	}
	return nil
}

func wsClient(serverURL *url.URL, onConnect func(*websocket.Conn), onClose func(), onTextMessage func([]byte), externalClose chan struct{}) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	log.Printf("connecting to %s", serverURL.String())

	c, _, err := websocket.DefaultDialer.Dial(serverURL.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Println("connected")

	defer c.Close()
	onConnect(c)
	defer onClose()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			msgType, message, err := c.ReadMessage()
			if err != nil {
				log.Println("error on read:", err)
				return
			}
			switch msgType {
			case websocket.BinaryMessage:
				//ignore for now
				log.Println("ignored binary message")
			case websocket.TextMessage:
				onTextMessage(message)
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-externalClose:
			close(interrupt)
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("error on write close:", err)
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
