/*
	Server for handle gloabal key for Yandex.Music service

bookmarklet:
	javascript:(function(){var js=document.createElement("script");document.body.appendChild(js);js.src='https://localhost:8900/script.js'})()

Generate sertificates:
	$ openssl genrsa -out server.key 2048
	$ openssl req -new -x509 -key server.key -out server.pem -days 3650

*/
package main

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

	var ws = new WebSocket("wss://localhost:8900/listen_keys");

	ws.onopen = function() {
		ws.send("Hello server"); 
	};

	ws.onerror = function() {
		console.log("Error open websocket")
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
	}
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

	for _, action := range [...]string{"pause", "prev", "next"} {
		action := action
		http.HandleFunc("/"+action, func(rw http.ResponseWriter, req *http.Request) {
			log.Printf("%s %s %s", req.Method, req.URL.Path, req.UserAgent())
			fmt.Fprint(rw, action+" ok")
			go func() {
				events <- action
			}()
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
		log.Print("Recived: ", string(message[:n]))

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
