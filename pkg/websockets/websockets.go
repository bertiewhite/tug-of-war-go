package websockets

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

type Websocket struct {
	u    url.URL
	conn *websocket.Conn
}

func NewWebsocketConnection(u url.URL) Websocket {
	return Websocket{
		u: u,
	}
}

func (w *Websocket) Dial() error {
	c, _, err := websocket.DefaultDialer.Dial(w.u.String(), nil)
	if err != nil {
		return err
	}
	w.conn = c
	return nil
}

var upgrader = websocket.Upgrader{}

func (w *Websocket) StartServer(action func(int, []byte) error, errChan chan<- error) error {
	handler := func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()

		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, 1000) {
					fmt.Println("Closing connection")
					return
				} else if websocket.IsUnexpectedCloseError(err, 1000) {
					fmt.Println("Unexpected close error", err.Error())
					return
				}
				errChan <- err
			}
			if messageType != websocket.TextMessage {
				fmt.Println("Some other message type: ", messageType)
			}

			err = action(messageType, message)
			if err != nil {
				go func() {
					errChan <- err
				}()
			}
		}
	}

	http.HandleFunc(w.u.Path, handler)
	err := http.ListenAndServe(w.u.Host, nil)

	return err
}

func (w *Websocket) Close() error {
	err := w.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "closing message"))
	return err
}

func (w *Websocket) SendJson(message any) error {

	json, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = w.conn.WriteMessage(websocket.TextMessage, json)
	if err != nil {
		return err
	}

	return nil
}
