package applicationsvc

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/bertiewhite/tug-of-war-go/pkg/websockets"
)

var u = url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/test"}

func Start() {
	be := os.Getenv("SENDER_RECIEVER")
	fmt.Printf(fmt.Sprintf("Starting %s\n", be))
	switch os.Getenv("SENDER_RECIEVER") {
	case "SENDER":
		launchSender()
	case "RECIEVER":
		launchReciever()
	default:
		fmt.Println("Could not find instruction")
	}
}

type DummyMessage struct {
	Hello     string `json:"hello"`
	SomeDummy string `json:"someDummy"`
	Count     int    `json:"count"`
}

func launchSender() {

	ws := websockets.NewWebsocketConnection(u)
	err := ws.Dial()
	if err != nil {
		fmt.Println("Error whilst dialing", err.Error())
		return
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ticker := time.NewTicker(1 * time.Second)
	count := 0
	for {
		select {
		case <-ticker.C:
			count++
			err := ws.SendJson(DummyMessage{
				Hello:     "World",
				SomeDummy: "I don't really know why I included this but hey ho I thought I should/could",
				Count:     count,
			})
			if err != nil {
				fmt.Printf("Error sending message wahhhhhhhhhhhhh: %e\n", err)
			}
		case <-interrupt:
			err := ws.Close()
			if err != nil {
				fmt.Printf("Error closing websocket: %s\n", err.Error())
			}
			fmt.Println("Closed channel")
			return
		}
	}
}

func launchReciever() {
	ws := websockets.NewWebsocketConnection(u)

	handler := func(i int, message []byte) error {
		rcved := DummyMessage{}

		err := json.Unmarshal(message, &rcved)
		if err != nil {
			fmt.Printf("Error unmarhalling json uh oh: %s\n", err.Error())
			// don't send this error to the error channel for now
			return nil
		}

		fmt.Printf("Message recieved: %+v\n", rcved)
		return nil
	}

	errChan := make(chan error)

	go func() {
		for err := range errChan {
			fmt.Printf("An error out of the error channel, that wasn't mean to happen yet: %s\n", err)
		}
	}()

	err := ws.StartServer(handler, errChan)
	close(errChan)
	log.Fatalf("error: %s", err.Error())
}
