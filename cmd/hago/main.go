package main

import (
	"flag"
	"net/url"

	"github.com/enolgor/hago"
	"github.com/enolgor/hago/homeassistant"
)

var mode string

const (
	modeFlag        = "mode"
	modeFlagShort   = "m"
	modeFlagDefault = ""
	modeFlagUsage   = "Specify mode (client/server)"
)

func init() {
	flag.StringVar(&mode, modeFlag, modeFlagDefault, modeFlagUsage)
	flag.StringVar(&mode, modeFlagShort, modeFlagDefault, modeFlagUsage+" (short)")
	flag.Parse()
}

func main() {
	url := &url.URL{Scheme: "wss", Host: "home.enolgor.es", Path: "/api/websocket"}
	wsconsummer := &homeassistant.HAWSConsummer{}
	switch mode {
	case "client":
		hago.WSClient(url, wsconsummer)
	default:
		flag.PrintDefaults()
	}
}
