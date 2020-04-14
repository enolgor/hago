package homeassistant

import "encoding/json"

// Constants

const HA_MSG_TYPE_AUTH_REQUIRED = "auth_required"
const HA_MSG_TYPE_AUTH = "auth"
const HA_MSG_TYPE_AUTH_OK = "auth_ok"
const HA_MSG_TYPE_AUTH_INVALID = "auth_invalid"
const HA_MSG_TYPE_RESULT = "result"
const HA_MSG_TYPE_SUBSCRIBE_EVENTS = "subscribe_events"
const HA_MSG_TYPE_EVENT = "event"

//Constructor mapper

// Home Assistant Generic Message

type HAMessage interface {
	GetType() string
}

type HAMessageImpl struct {
	Type string `json:"type"`
}

func (hm *HAMessageImpl) GetType() string {
	return hm.Type
}

// Auth Required Message

type AuthRequiredMessage interface {
	HAMessage
}

type authRequiredMessageImpl struct {
	*HAMessageImpl
}

func NewAuthRequiredMessage() AuthRequiredMessage {
	authRequiredMessage := &authRequiredMessageImpl{HAMessageImpl: &HAMessageImpl{}}
	authRequiredMessage.Type = HA_MSG_TYPE_AUTH_REQUIRED
	return authRequiredMessage
}

// Auth Message

type AuthMessage interface {
	HAMessage
	GetAccessToken() string
}

type authMessageImpl struct {
	*HAMessageImpl
	AccessToken string `json:"access_token"`
}

func (am *authMessageImpl) GetAccessToken() string {
	return am.AccessToken
}

func NewAuthMessage(accessToken string) AuthMessage {
	authMessage := &authMessageImpl{HAMessageImpl: &HAMessageImpl{}}
	authMessage.Type = HA_MSG_TYPE_AUTH
	authMessage.AccessToken = accessToken
	return authMessage
}

// Auth Ok Message

type AuthOkMessage interface {
	HAMessage
}

type authOkMessageImpl struct {
	*HAMessageImpl
}

func NewAuthOkMessage() AuthOkMessage {
	authOkMessage := &authOkMessageImpl{HAMessageImpl: &HAMessageImpl{}}
	authOkMessage.Type = HA_MSG_TYPE_AUTH_OK
	return authOkMessage
}

// Auth Invalid Message

type AuthInvalidMessage interface {
	HAMessage
	GetMessage() string
}

type authInvalidMessageImpl struct {
	*HAMessageImpl
	Message string `json:"message"`
}

func (aim *authInvalidMessageImpl) GetMessage() string {
	return aim.Message
}

func NewAuthInvalidMessage(message string) AuthInvalidMessage {
	authInvalidMessage := &authInvalidMessageImpl{HAMessageImpl: &HAMessageImpl{}}
	authInvalidMessage.Type = HA_MSG_TYPE_AUTH_INVALID
	authInvalidMessage.Message = message
	return authInvalidMessage
}

// Identified Message

type IdentifiedMessage interface {
	HAMessage
	GetId() int
}

type IdentifiedMessageImpl struct {
	*HAMessageImpl
	Id int `json:"id"`
}

func (im *IdentifiedMessageImpl) GetId() int {
	return im.Id
}

// Subscribe to events

type SubscribeEventsMessage interface {
	IdentifiedMessage
	GetEventType() string
}

type subscribeEventsMessageImpl struct {
	*IdentifiedMessageImpl
	EventType string `json:"event_type,omitempty"`
}

func (semi *subscribeEventsMessageImpl) GetEventType() string {
	return semi.EventType
}

func NewSubscribeEventsMessage(id int, eventType string) SubscribeEventsMessage {
	subscribeEventsMessage := &subscribeEventsMessageImpl{IdentifiedMessageImpl: &IdentifiedMessageImpl{HAMessageImpl: &HAMessageImpl{}}}
	subscribeEventsMessage.Type = HA_MSG_TYPE_SUBSCRIBE_EVENTS
	subscribeEventsMessage.Id = id
	subscribeEventsMessage.EventType = eventType
	return subscribeEventsMessage
}

// Result Message

type ResultMessage interface {
	IdentifiedMessage
	GetSuccess() bool
}

type resultMessageImpl struct {
	*IdentifiedMessageImpl
	Success bool `json:"success"`
}

func (rmi *resultMessageImpl) GetSuccess() bool {
	return rmi.Success
}

func NewResultMessage(id int, success bool) ResultMessage {
	resultMessage := &resultMessageImpl{IdentifiedMessageImpl: &IdentifiedMessageImpl{HAMessageImpl: &HAMessageImpl{}}}
	resultMessage.Type = HA_MSG_TYPE_RESULT
	resultMessage.Id = id
	resultMessage.Success = success
	return resultMessage
}

// Event Message

type EventMessage interface {
	IdentifiedMessage
	GetEvent() *map[string]interface{}
}

type eventMessageImpl struct {
	*IdentifiedMessageImpl
	Event *map[string]interface{} `json:"event"`
}

func (emi *eventMessageImpl) GetEvent() *map[string]interface{} {
	return emi.Event
}

func NewEventMessage(id int, event *map[string]interface{}) EventMessage {
	eventMessageImpl := &eventMessageImpl{IdentifiedMessageImpl: &IdentifiedMessageImpl{HAMessageImpl: &HAMessageImpl{}}}
	eventMessageImpl.Type = HA_MSG_TYPE_EVENT
	eventMessageImpl.Id = id
	eventMessageImpl.Event = event
	return eventMessageImpl
}

// Unmarshall All

func ParseMessage(data []byte) (HAMessage, error) {
	var obj map[string]interface{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}
	msgType := ""
	if t, ok := obj["type"].(string); ok {
		msgType = t
	}
	var actual HAMessage
	switch msgType {
	case HA_MSG_TYPE_AUTH_REQUIRED:
		actual = &authRequiredMessageImpl{}
	case HA_MSG_TYPE_AUTH:
		actual = &authMessageImpl{}
	case HA_MSG_TYPE_AUTH_OK:
		actual = &authOkMessageImpl{}
	case HA_MSG_TYPE_AUTH_INVALID:
		actual = &authInvalidMessageImpl{}
	case HA_MSG_TYPE_RESULT:
		actual = &resultMessageImpl{}
	case HA_MSG_TYPE_SUBSCRIBE_EVENTS:
		actual = &subscribeEventsMessageImpl{}
	}
	err = json.Unmarshal(data, actual)
	if err != nil {
		return nil, err
	}
	return actual, nil
}
