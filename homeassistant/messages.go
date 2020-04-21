package homeassistant

import (
	"encoding/json"
	"fmt"
	"time"
)

const HA_MSG_TYPE_AUTH_REQUIRED = "auth_required"
const HA_MSG_TYPE_AUTH = "auth"
const HA_MSG_TYPE_AUTH_OK = "auth_ok"
const HA_MSG_TYPE_AUTH_INVALID = "auth_invalid"
const HA_MSG_TYPE_RESULT = "result"
const HA_MSG_TYPE_SUBSCRIBE_EVENTS = "subscribe_events"
const HA_MSG_TYPE_EVENT = "event"
const HA_MSG_TYPE_CALL_SERVICE = "call_service"
const HA_MSG_TYPE_GET_STATES = "get_states"

type Message interface {
	GetType() string
	setType(messageType string)
	GetID() int
	SetID(id int)
}

type message struct {
	Type string `json:"type"`
	ID   *int   `json:"id,omitempty"`
}

func (m *message) GetType() string {
	return m.Type
}

func (m *message) setType(messageType string) {
	m.Type = messageType
}

func (m *message) GetID() int {
	return *m.ID
}

func (m *message) SetID(id int) {
	m.ID = &id
}

type AuthRequiredMessage struct {
	message
}

type AuthMessage struct {
	message
	AccessToken string `json:"access_token"`
}

type AuthOkMessage struct {
	message
}

type AuthInvalidMessage struct {
	message
}

type SubscribeEventsMessage struct {
	message
	EventType *string `json:"event_type,omitempty"`
}

type ResultMessage struct {
	message
	Success bool    `json:"success"`
	Error   *Error  `json:"error,omitempty"`
	Result  *Result `json:"result,omitempty"`
}

type Result struct {
	States []State
	Map    map[string]interface{}
}

type State struct {
	Attributes  map[string]interface{} `json:"attributes"`
	EntityID    string                 `json:"entity_id"`
	LastChanged EventTime              `json:"last_changed"`
	LastUpdated EventTime              `json:"last_updated"`
	State       string                 `json:"state"`
}

func (r *Result) UnmarshalJSON(data []byte) error {
	if data[0] == '[' {
		return json.Unmarshal(data, &r.States)
	}
	if data[0] == '{' {
		return json.Unmarshal(data, &r.Map)
	}
	return fmt.Errorf("Result should be object or array")
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var knownErrorCodes map[int]string = map[int]string{
	1: "A non-increasing identifier has been supplied.",
	2: "Received message is not in expected format (voluptuous validation error).",
	3: "Requested item cannot be found.",
}

func (e *Error) Error() string {
	desc, ok := knownErrorCodes[e.Code]
	if !ok {
		desc = "Unknown error code."
	}
	return fmt.Sprintf("Code %d. %s. %s", e.Code, desc, e.Message)
}

type EventMessage struct {
	message
	Event *Event `json:"event"`
}

type commonEvent struct {
	EventType string     `json:"event_type"`
	TimeFired *EventTime `json:"time_fired"`
	Origin    string     `json:"origin"`
	Data      map[string]interface{}
}

type Event struct {
	commonEvent
	StateChangedData *StateChangedData
}

type StateChangedData struct {
	EntityID string `json:"entity_id"`
	NewState *State `json:"new_state"`
	OldState *State `json:"old_state"`
}

func (e *Event) UnmarshalJSON(data []byte) (err error) {
	err = json.Unmarshal(data, &e.commonEvent)
	if err != nil {
		return
	}
	switch e.EventType {
	case "state_changed":
		v := struct {
			StateChangedData *StateChangedData `json:"data"`
		}{}
		err = json.Unmarshal(data, &v)
		if err != nil {
			return
		}
		e.StateChangedData = v.StateChangedData
	}
	return
}

type EventTime struct {
	Value time.Time
}

type CallServiceMessage struct {
	message
	Domain  string                 `json:"domain"`
	Service string                 `json:"service"`
	Data    map[string]interface{} `json:"service_data,omitempty"`
}

type GetStatesMessage struct {
	message
}

func (e *EventTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, e.Value.Format("2006-01-02T15:04:05.000000-07:00"))), nil
}

func (e *EventTime) UnmarshalJSON(data []byte) error {
	parsed, err := time.Parse(`"2006-01-02T15:04:05.000000-07:00"`, string(data))
	if err != nil {
		return err
	}
	e.Value = parsed
	return nil
}

func NewMessage(messageType string) Message {
	var dst Message
	switch messageType {
	case HA_MSG_TYPE_AUTH_REQUIRED:
		dst = new(AuthRequiredMessage)
	case HA_MSG_TYPE_AUTH:
		dst = new(AuthMessage)
	case HA_MSG_TYPE_AUTH_OK:
		dst = new(AuthOkMessage)
	case HA_MSG_TYPE_AUTH_INVALID:
		dst = new(AuthInvalidMessage)
	case HA_MSG_TYPE_RESULT:
		dst = new(ResultMessage)
	case HA_MSG_TYPE_SUBSCRIBE_EVENTS:
		dst = new(SubscribeEventsMessage)
	case HA_MSG_TYPE_EVENT:
		dst = new(EventMessage)
	case HA_MSG_TYPE_CALL_SERVICE:
		dst = new(CallServiceMessage)
	case HA_MSG_TYPE_GET_STATES:
		dst = new(GetStatesMessage)
	}
	dst.setType(messageType)
	return dst
}

func ParseMessage(data []byte) (Message, error) {
	var m message
	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	dst := NewMessage(m.Type)
	if m.ID != nil {
		dst.SetID(*m.ID)
	}
	json.Unmarshal(data, &dst)
	return dst, nil
}
