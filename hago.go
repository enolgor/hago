package hago

import (
	"github.com/enolgor/hago/homeassistant"
)

type ListenerHandler uint16

type Client interface {
	Connect(onConnect func(), onClose func())
	SubscribeToEvent(eventType string, onEvent func(*homeassistant.Event)) ListenerHandler
	SubscribeToState(entityID string, onStateChange func(*homeassistant.StateChangedData)) ListenerHandler
	UnsubscribeFromEvent(handler ListenerHandler)
	UnsubscribeFromState(handler ListenerHandler)
	CallService(domain string, service string, data map[string]interface{}) error
	FetchStates() (map[string]homeassistant.State, error)
	Close()
}
