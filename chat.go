package goroomchat

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/novalagung/gubrak"
)

type M map[string]interface{}

const MESSAGE_NEW_USER = "New User"
const MESSAGE_CHAT = "Chat"
const MESSAGE_LEAVE = "Leave"

// var connections = make([]*WebSocketConnection, 0)
var rooms = make(map[string][]*WebSocketConnection)

type SocketPayload struct {
	Message string
}

type SocketResponse struct {
	From    string
	Type    string
	Message string
}

type WebSocketConnection struct {
	*websocket.Conn
	Username string
}

func HandleWS(w http.ResponseWriter, r *http.Request) {
	currentGorillaConn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	username := r.URL.Query().Get("username")
	room := r.URL.Query().Get("room")

	currentConn := WebSocketConnection{Conn: currentGorillaConn, Username: username}
	rooms[room] = append(rooms[room], &currentConn)

	go handleIO(&currentConn, room)
}

func HandleRooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		result, err := json.Marshal(rooms)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(result)
		return
	}
	http.Error(w, "", http.StatusBadRequest)
}

func handleIO(currentConn *WebSocketConnection, room string) {

	defer func() {
		if r := recover(); r != nil {
			log.Println("ERROR", fmt.Sprintf("%v", r))
		}
	}()

	broadcastMessage(currentConn, room, MESSAGE_NEW_USER, "")

	for {
		payload := SocketPayload{}
		err := currentConn.ReadJSON(&payload)
		if err != nil {
			if strings.Contains(err.Error(), "websocket: close") {
				broadcastMessage(currentConn, room, MESSAGE_LEAVE, "")
				ejectConnection(currentConn, room)
				return
			}

			log.Println("ERROR", err.Error())
			continue
		}

		broadcastMessage(currentConn, room, MESSAGE_CHAT, payload.Message)
	}
}

func ejectConnection(currentConn *WebSocketConnection, room string) {
	filtered, _ := gubrak.Reject(rooms[room], func(each *WebSocketConnection) bool {
		return each == currentConn
	})
	rooms[room] = filtered.([]*WebSocketConnection)
	if len(rooms[room]) == 0 {
		delete(rooms, room)
	}
}

func broadcastMessage(currentConn *WebSocketConnection, room string, kind, message string) {
	for _, eachConn := range rooms[room] {
		if eachConn == currentConn {
			continue
		}

		eachConn.WriteJSON(SocketResponse{
			From:    currentConn.Username,
			Type:    kind,
			Message: message,
		})
	}
}
