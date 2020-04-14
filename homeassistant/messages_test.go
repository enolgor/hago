package homeassistant

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"unicode"
)

func TestMessages(t *testing.T) {
	authRequiredMessage := NewAuthRequiredMessage()
	want := "auth_required"
	got := authRequiredMessage.GetType()
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func StripWhitespace(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func TestMarshalMessages(t *testing.T) {
	messages := []struct {
		name    string
		msg     HAMessage
		jsonMsg string
	}{
		{
			name: "AuthRequiredMessage",
			msg:  NewAuthRequiredMessage(),
			jsonMsg: `
				{ "type": "auth_required" }
			`,
		},
		{
			name: "AuthMessage",
			msg:  NewAuthMessage("123"),
			jsonMsg: `
				{ "type": "auth", "access_token": "123" }
			`,
		},
		{
			name: "AuthOkMessage",
			msg:  NewAuthOkMessage(),
			jsonMsg: `
				{ "type": "auth_ok" }
			`,
		},
		{
			name: "AuthInvalidMessage",
			msg:  NewAuthInvalidMessage("Invalidpassword"),
			jsonMsg: `
				{ "type": "auth_invalid", "message": "Invalidpassword" }
			`,
		},
		{
			name: "ResultMessage",
			msg:  NewResultMessage(13, false),
			jsonMsg: `
				{ "type": "result", "id": 13, "success": false }
			`,
		},
		{
			name: "SubscribeEventsMessage",
			msg:  NewSubscribeEventsMessage(13, "state_changed"),
			jsonMsg: `
				{ "type": "subscribe_events", "id": 13, "event_type": "state_changed" }
			`,
		},
		{
			name: "EventMessage",
			msg:  NewEventMessage(13, nil),
			jsonMsg: `
				{ "type": "event", "id": 13, "event": null }
			`,
		},
	}
	testMarshall := func(t *testing.T, message HAMessage, jsonMessage string) {
		t.Helper()
		got, _ := json.Marshal(message)
		want := StripWhitespace(jsonMessage)
		if string(got) != StripWhitespace(jsonMessage) {
			t.Errorf("got %q want %q", string(got), want)
		}
	}
	testUnMarshall := func(t *testing.T, message HAMessage, jsonMessage string) {
		t.Helper()
		got, err := ParseMessage([]byte(jsonMessage))
		if err != nil {
			t.Errorf("On Parse Message: %s", err)
		}
		want := message
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %q want %q", got, want)
		}
	}
	for _, item := range messages {
		t.Run(item.name, func(t *testing.T) {
			testMarshall(t, item.msg, item.jsonMsg)
			testUnMarshall(t, item.msg, item.jsonMsg)
		})
	}
}
