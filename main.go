package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pelletier/go-toml"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// client list
var webClients = make(map[int]*websocket.Conn)

func watchServer(w http.ResponseWriter, r *http.Request) {
	// block until stream is closed
	defer func(Body io.ReadCloser) {
		err := Body.Close()

		if err != nil {
			fmt.Println("Failed to close HTTP request-body stream")
		}
	}(r.Body)

	data, err := io.ReadAll(r.Body)

	if err != nil {
		fmt.Println("Failure while reading request-body")
		return
	}

	for _, client := range webClients {
		err := client.WriteMessage(websocket.TextMessage, data)

		if err != nil {
			fmt.Println("Failed to write data to websocket")
		}
	}
}

func webServer(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	var index = len(webClients) + 1
	fmt.Printf("Client #%d connected\n", index)

	webClients[index] = ws

	for {
		/*
			block until a message is sent (shouldn't ever be sent)
			an error is automatically thrown if the websocket cleanly closes

			i.e. OBS closed or the browser tab was closed
		*/
		_, _, err = ws.ReadMessage()

		if err != nil {
			fmt.Printf("Client #%d disconnected\n", index)

			// clip the socket out, something went wrong (disconnected?)
			delete(webClients, index)
			break
		}
	}
}

func main() {
	var err error = nil

	var config *toml.Tree
	config, err = toml.LoadFile("settings.toml")

	if err != nil {
		fmt.Println(err)
		time.Sleep(5000)

		os.Exit(1)
	}

	var path = config.Get("file.path").(string)

	var fileServer = http.FileServer(http.Dir(path))

	// let the content server run
	http.Handle("/", fileServer)

	// handle the Apple Watch
	http.HandleFunc("/watch", watchServer)

	// handle the web clients
	http.HandleFunc("/web", webServer)

	fmt.Println("Successfully routed websocket paths")

	var port = uint16(config.Get("server.port").(int64)) // can't use uint16
	var ip = config.Get("server.address").(string)

	var address = fmt.Sprintf("%s:%d", ip, port)

	// print the workingDirectory serving from and the address
	var workingDirectory string
	workingDirectory, err = os.Getwd()

	if err != nil {
		fmt.Println(err)
		time.Sleep(5000)

		os.Exit(1)
	}

	// make sure to append "path"
	var fullPath = fmt.Sprintf("%s\\%s", workingDirectory, path)
	fmt.Printf("Serving content on \"%s\" from \"%s\"\n", address, fullPath)
	fmt.Print("\n")

	// block the main thread and serve http requests
	err = http.ListenAndServe(address, nil)

	if err != nil {
		fmt.Println(err)
		time.Sleep(5000)

		os.Exit(1)
	}
}
