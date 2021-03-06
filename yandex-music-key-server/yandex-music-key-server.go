package main

/*
Server for handle global key for Yandex.Music service

bookmarklet:
	javascript:(function(){var js=document.createElement("script");document.body.appendChild(js);js.src='https://localhost:8900/script.js'})()

Generate sertificates:
	$ openssl genrsa -out server.key 2048
	$ openssl req -new -x509 -key server.key -out server.pem -days 3650

or convert from .p12 (exported from keychain):
	$ openssl pkcs12 -in cert.p12 -out server.pem -nodes

Hammerspoon config for global keys:
-- Yandex.Music hotkeys via server
hs.hotkey.bind({"cmd", "alt", "ctrl"}, "F8", function()
    hs.http.get("https://localhost:8900/pause", nil)
end)
hs.hotkey.bind({"cmd", "alt", "ctrl"}, "F7", function()
    hs.http.get("https://localhost:8900/prev", nil)
end)
hs.hotkey.bind({"cmd", "alt", "ctrl"}, "F9", function()
    hs.http.get("https://localhost:8900/next", nil)
end)

Download from:
	https://github.com/msoap/etc/tree/master/yandex-music-key-server

*/

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

// TODO auto detect free port
const (
	PORT  = "8900"
	INDEX = `
<!doctype html>
<head>
	<meta charset="utf-8">
	<title>Yandex Music Key Server</title>
	<script src="/script.js"></script>
</head>
<body>
	<h1>Yandex Music Key Server</h1>  
</body>
</html>
	`
	JS = `
(function () {
	console.log("Hello ya.music")

	var connect_ws = function () {
		console.log("attempt connect")
		var ws = new WebSocket("wss://localhost:` + PORT + `/listen_keys");

		ws.onopen = function() {
			console.log("Websocket connected")
			ws.send("Hello server"); 
		};

		ws.onerror = function() {
			console.log("Error open websocket")
		};

		ws.onclose = function () {
			console.log("WS connect close")
			setTimeout(function () {
				connect_ws();
			}, 5 * 1000);
		};

		ws.onmessage = function(event) {
			console.log("Message from server:", event.data);
			switch (event.data) {
				case "pause":
					$('.player-controls__btn_play').click();
					break;
				case "prev":
					$('.player-controls__btn_prev').click();
					break;
				case "next":
					$('.player-controls__btn_next').click();
					break;
				default:
					console.log("Action not found: " + event.data);
			}
		};
	};
	connect_ws();
})()
	`
)

func main() {
	log.Print("yandex-music-key-server start")
	events := make(chan string)

	// Setup http server
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s %s", req.Method, req.URL.Path, req.UserAgent())
		fmt.Fprint(rw, INDEX)
	})
	http.HandleFunc("/script.js", func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s %s", req.Method, req.URL.Path, req.UserAgent())
		fmt.Fprint(rw, JS)
	})

	websocketConnected := false

	for _, action := range [...]string{"pause", "prev", "next"} {
		action := action
		http.HandleFunc("/"+action, func(rw http.ResponseWriter, req *http.Request) {
			log.Printf("%s %s %s", req.Method, req.URL.Path, req.UserAgent())
			fmt.Fprint(rw, action+" ok")
			if websocketConnected {
				go func() {
					events <- action
				}()
			} else {
				log.Print("websocket isn't connected")
			}
		})
	}

	// Setup websocket server
	http.Handle("/listen_keys", websocket.Handler(func(ws *websocket.Conn) {
		log.Print("WebSocket connect")
		message := make([]byte, 100)
		n, err := ws.Read(message)
		if err != nil {
			log.Print(err)
			return
		}
		log.Print("Received: ", string(message[:n]))

		websocketConnected = true
		defer func() {
			websocketConnected = false
		}()

		for {
			select {
			case event := <-events:
				_, err := io.WriteString(ws, event)
				if err != nil {
					log.Print("Send error: ", err)
					return
				}
				log.Print("Send key: " + event)
			}
		}
	}))

	log.Print("Listen websocket on https://localhost:" + PORT)
	err := http.ListenAndServeTLS("localhost:"+PORT, "server.pem", "server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServeTLS: " + err.Error())
	}
}
