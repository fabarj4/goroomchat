package goroomchat

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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
	Room    string
	To      string
}

type SocketResponse struct {
	From    string
	Type    string
	Message string
	To      string
}

type WebSocketConnection struct {
	*websocket.Conn
	Username string
}

type Rooms struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func GetRooms() []*Rooms {
	var result []*Rooms
	for roomName, _ := range rooms {
		temp := &Rooms{Name: roomName, Status: "online"}
		result = append(result, temp)
	}
	return result
}

func HandleRooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

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

func HandleWS(w http.ResponseWriter, r *http.Request) {
	currentGorillaConn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	username := r.URL.Query().Get("username")
	room := r.URL.Query().Get("room")
	to := r.URL.Query().Get("to")

	fmt.Println(to)

	currentConn := WebSocketConnection{Conn: currentGorillaConn, Username: username}
	rooms[room] = append(rooms[room], &currentConn)

	go handleIO(&currentConn, room)
}

func handleIO(currentConn *WebSocketConnection, room string) {

	defer func() {
		if r := recover(); r != nil {
			log.Println("ERROR", fmt.Sprintf("%v", r))
		}
	}()

	//buat log pertama kali untuk user
	if err := createLog(currentConn.Username); err != nil {
		fmt.Println(err.Error())
		return
	}
	for {
		payload := SocketPayload{}
		err := currentConn.ReadJSON(&payload)
		if err != nil {
			if strings.Contains(err.Error(), "websocket: close") {
				broadcastMessage(currentConn, room, payload.To, MESSAGE_LEAVE, "")
				ejectConnection(currentConn, room)
				return
			}

			log.Println("ERROR", err.Error())
			continue
		}

		//memasukan chat ke dalam log file
		if err := writeToLog(currentConn.Username, payload.To, payload.Message); err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Println(payload)

		broadcastMessage(currentConn, room, payload.To, MESSAGE_CHAT, payload.Message)
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

func broadcastMessage(currentConn *WebSocketConnection, room, to, kind, message string) {

	for _, eachConn := range rooms[room] {
		if eachConn == currentConn {
			continue
		}

		eachConn.WriteJSON(SocketResponse{
			From:    currentConn.Username,
			Type:    kind,
			Message: message,
			To:      to,
		})

	}
}

func createLog(username string) error {
	folderName := "log"

	_, err := os.Stat(folderName)
	if os.IsNotExist(err) {
		if err := os.Mkdir(folderName, os.FileMode(0644)); err != nil {
			return err
		}
	}

	path := fmt.Sprintf("%s/%s.txt", folderName, username)
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
	}
	return nil
}

func writeToLog(from, to, message string) error {
	//pencatatan log pengirim
	logFile, err := os.OpenFile(fmt.Sprintf("log/%s.txt", from), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	messageData := fmt.Sprintf(`{"from":"%s", "type":"%s", "message":"%s", "to":"%s"}`, from, MESSAGE_CHAT, message, to)
	if _, err := logFile.WriteString(fmt.Sprintf("%v\n", messageData)); err != nil {
		return err
	}
	if err := logFile.Close(); err != nil {
		return err
	}
	//pencatatan log penerima
	logFile, err = os.OpenFile(fmt.Sprintf("log/%s.txt", to), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	messageData = fmt.Sprintf(`{"from":"%s", "type":"%s", "message":"%s", "to":"%s"}`, from, MESSAGE_CHAT, message, to)
	if _, err := logFile.WriteString(fmt.Sprintf("%v\n", messageData)); err != nil {
		return err
	}
	if err := logFile.Close(); err != nil {
		return err
	}
	// path := fmt.Sprintf("log/%s.txt", from)
	// file, err := os.OpenFile(path, os.O_RDWR, 0644)
	// if err != nil {
	// 	return err
	// }
	//
	// var text = make([]byte, 1024)
	// for {
	// 	n, err := file.Read(text)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if err != io.EOF {
	// 		if err != nil {
	// 			break
	// 		}
	// 	}
	// 	if n == 0 {
	// 		break
	// 	}
	// }
	// fmt.Println(text)
	// messageData := fmt.Sprintf(`{"from":"%s", "type":"%s", "message":"%s", "to":"%s"}`, from, MESSAGE_CHAT, message, to)
	// if _, err := file.WriteString(messageData); err != nil {
	// 	return err
	// }
	// if err := file.Close(); err != nil {
	// 	return err
	// }
	//
	//pencatatan log penerima
	// path = fmt.Sprintf("log/%s.txt", to)
	// file, err = os.OpenFile(path, os.O_RDWR, 0644)
	// if err != nil {
	// 	return err
	// }
	// messageData = fmt.Sprintf(`{"from":"%s", "type":"%s", "message":"%s", "to":"%s"}`, to, MESSAGE_CHAT, message, from)
	// if _, err := file.WriteString(messageData); err != nil {
	// 	return err
	// }
	// if err := file.Close(); err != nil {
	// 	return err
	// }

	return nil
}
