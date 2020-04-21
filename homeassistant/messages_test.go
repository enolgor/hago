package homeassistant

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestMarshaling(t *testing.T) {
	authRequiredMessage := NewMessage(HA_MSG_TYPE_AUTH_REQUIRED).(*AuthRequiredMessage)
	authRequiredMessageJSON := `{"type":"auth_required"}`

	authMessage := NewMessage(HA_MSG_TYPE_AUTH).(*AuthMessage)
	authMessage.AccessToken = "123"
	authMessageJSON := `{"type":"auth","access_token":"123"}`

	authOkMessage := NewMessage(HA_MSG_TYPE_AUTH_OK).(*AuthOkMessage)
	authOkMessageJSON := `{"type":"auth_ok"}`

	authInvalidMessage := NewMessage(HA_MSG_TYPE_AUTH_INVALID).(*AuthInvalidMessage)
	authInvalidMessageJSON := `{"type":"auth_invalid"}`

	resultMessage := NewMessage(HA_MSG_TYPE_RESULT).(*ResultMessage)
	resultMessage.SetID(13)
	resultMessage.Success = true
	resultMessage.Error = nil
	resultMessageJSON := `{"type":"result","id":13,"success":true}`

	errorMessage := NewMessage(HA_MSG_TYPE_RESULT).(*ResultMessage)
	errorMessage.SetID(13)
	errorMessage.Success = false
	errorMessage.Error = &Error{Code: 123, Message: "some_error_message"}
	errorMessageJSON := `{"type":"result","id":13,"success":false,"error":{"code":123,"message":"some_error_message"}}`

	subscribeEventsMessage := NewMessage(HA_MSG_TYPE_SUBSCRIBE_EVENTS).(*SubscribeEventsMessage)
	subscribeEventsMessage.SetID(12)
	eventType := "some_event"
	subscribeEventsMessage.EventType = &eventType
	subscribeEventsMessageJSON := `{"type":"subscribe_events","id":12,"event_type":"some_event"}`

	subscribeAllEventsMessage := NewMessage(HA_MSG_TYPE_SUBSCRIBE_EVENTS).(*SubscribeEventsMessage)
	subscribeAllEventsMessage.SetID(12)
	subscribeAllEventsMessage.EventType = nil
	subscribeAllEventsMessageJSON := `{"type":"subscribe_events","id":12}`

	eventMessage := NewMessage(HA_MSG_TYPE_EVENT).(*EventMessage)
	eventMessage.SetID(18)
	eventMessage.Event = &Event{}
	eventMessage.Event.EventType = "state_changed"
	timeFired, _ := time.Parse(time.RFC3339, "2016-11-26T01:37:24.265429+00:00")
	eventMessage.Event.TimeFired = &EventTime{timeFired}
	eventMessage.Event.Origin = "LOCAL"
	eventMessage.Event.Data = map[string]interface{}{
		"entity_id": "light.bed_light",
		"new_state": map[string]interface{}{
			"entity_id":    "light.bed_light",
			"last_changed": "2016-11-26T01:37:24.265390+00:00",
			"last_updated": "2016-11-26T01:37:24.265390+00:00",
			"state":        "on",
		},
		"old_state": map[string]interface{}{
			"entity_id":    "light.bed_light",
			"last_changed": "2016-11-26T01:37:10.466994+00:00",
			"last_updated": "2016-11-26T01:37:10.466994+00:00",
			"state":        "off",
		},
	}

	eventMessageJSON := `{"type":"event","id":18,"event":{"data":{"entity_id":"light.bed_light","new_state":{"entity_id":"light.bed_light","last_changed":"2016-11-26T01:37:24.265390+00:00","last_updated":"2016-11-26T01:37:24.265390+00:00","state":"on"},"old_state":{"entity_id":"light.bed_light","last_changed":"2016-11-26T01:37:10.466994+00:00","last_updated":"2016-11-26T01:37:10.466994+00:00","state":"off"}},"event_type":"state_changed","time_fired":"2016-11-26T01:37:24.265429+00:00","origin":"LOCAL"}}`

	messageTests := []struct {
		name    string
		msg     Message
		jsonMsg string
	}{
		{
			name:    "AuthRequiredMessage",
			msg:     authRequiredMessage,
			jsonMsg: authRequiredMessageJSON,
		},
		{
			name:    "AuthMessage",
			msg:     authMessage,
			jsonMsg: authMessageJSON,
		},
		{
			name:    "AuthOkMessage",
			msg:     authOkMessage,
			jsonMsg: authOkMessageJSON,
		},
		{
			name:    "AuthInvalidMessage",
			msg:     authInvalidMessage,
			jsonMsg: authInvalidMessageJSON,
		},
		{
			name:    "ResultMessage",
			msg:     resultMessage,
			jsonMsg: resultMessageJSON,
		},
		{
			name:    "ErrorMessage",
			msg:     errorMessage,
			jsonMsg: errorMessageJSON,
		},
		{
			name:    "SubscribeEventsMessage",
			msg:     subscribeEventsMessage,
			jsonMsg: subscribeEventsMessageJSON,
		},
		{
			name:    "SubscribeAllEventsMessage",
			msg:     subscribeAllEventsMessage,
			jsonMsg: subscribeAllEventsMessageJSON,
		},
		{
			name:    "EventMessage",
			msg:     eventMessage,
			jsonMsg: eventMessageJSON,
		},
	}
	testMarshall := func(t *testing.T, message Message, jsonMessage string) {
		t.Helper()
		got, _ := json.Marshal(message)
		if string(got) != jsonMessage {
			t.Errorf("got %q want %q", string(got), jsonMessage)
		}
	}
	testUnMarshall := func(t *testing.T, message Message, jsonMessage string) {
		t.Helper()
		got, err := ParseMessage([]byte(jsonMessage))
		if err != nil {
			t.Errorf("On Parse Message: %s", err)
		}
		if !reflect.DeepEqual(got, message) {
			t.Errorf("got %q want %q", got, message)
		}
	}
	for _, item := range messageTests {
		t.Run(item.name, func(t *testing.T) {
			testMarshall(t, item.msg, item.jsonMsg)
			testUnMarshall(t, item.msg, item.jsonMsg)
		})
	}
}
