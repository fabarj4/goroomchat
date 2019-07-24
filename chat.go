package goroomchat

import (
	"bufio"
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

const MESSAGE_NEW_LOGIN = "new_login"
const MESSAGE_CHAT = "chat"
const MESSAGE_LEAVE = "leave"
const MESSAGE_LOG = "log"

// var connections = make([]*WebSocketConnection, 0)
var rooms = make(map[string][]*WebSocketConnection)

type SocketPayload struct {
	Message string `json:"message"`
	Room    string `json:"room"`
	To      string `json:"to"`
}

type SocketResponse struct {
	User    string `json:"user"`
	From    string `json:"from"`
	Type    string `json:"type"`
	Message string `json:"message"`
	Room    string `json:"room"`
	To      string `json:"to"`
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
	// to := r.URL.Query().Get("to")

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

	//ambil data dari log
	nameFile := currentConn.Username
	dataLogString, err := readLog(nameFile)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	data := []SocketResponse{}
	if err := json.Unmarshal([]byte(dataLogString), &data); err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, item := range data {
		broadcastMessage(currentConn.Username, item.Room, item.From, item.To, MESSAGE_LOG, item.Message)
	}

	for {
		payload := SocketPayload{}
		err := currentConn.ReadJSON(&payload)
		if err != nil {
			if strings.Contains(err.Error(), "websocket: close") {
				broadcastMessage(currentConn.Username, room, currentConn.Username, payload.To, MESSAGE_LEAVE, "")
				ejectConnection(currentConn, room)
				return
			}

			log.Println("ERROR", err.Error())
			continue
		}
		//pencatatan ke dalam log
		if err := writeLog(currentConn.Username, payload.To, payload.Room, payload.Message); err != nil {
			return
		}
		broadcastMessage(currentConn.Username, room, currentConn.Username, payload.To, MESSAGE_CHAT, payload.Message)
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

func broadcastMessage(user, room, from, to, kind, message string) {
	for _, eachConn := range rooms[room] {
		if from == to && kind != MESSAGE_LOG {
			// fmt.Printf("%v = %v", eachConn, currentConn)
			continue
		}

		//pencatatan kedalam log
		// if kind != MESSAGE_LOG && message != "" {
		// 	if room == "" && (from == eachConn.Username || to == eachConn.Username) {
		// 		if err := writeLog(eachConn.Username, to, "", message); err != nil {
		// 			return
		// 		}
		// 	} else {
		// 		if err := writeLog(eachConn.Username, "", room, message); err != nil {
		// 			return
		// 		}
		// 	}
		// }

		eachConn.WriteJSON(SocketResponse{
			User:    user,
			From:    from,
			Type:    kind,
			Message: message,
			Room:    room,
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

func writeLog(from, to, room, message string) error {
	// logFile, err := os.OpenFile(fmt.Sprintf("log/%s.txt", from), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	return err
	// }
	// messageData := fmt.Sprintf(`{"from":"%s", "type":"%s", "message":"%s", "room":"%s", "to":"%s"}`, from, MESSAGE_CHAT, message, room, to)
	// if _, err := logFile.WriteString(fmt.Sprintf("%v\n", messageData)); err != nil {
	// 	return err
	// }
	// if err := logFile.Close(); err != nil {
	// 	return err
	// }
	if to != "" {
		//pencatatan log pengirim
		logFile, err := os.OpenFile(fmt.Sprintf("log/%s.txt", from), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		messageData := fmt.Sprintf(`{"from":"%s", "type":"%s", "message":"%s", "room":"%s", "to":"%s"}`, from, MESSAGE_CHAT, message, room, to)
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
		messageData = fmt.Sprintf(`{"from":"%s", "type":"%s", "message":"%s", "room":"%s", "to":"%s"}`, from, MESSAGE_CHAT, message, room, to)
		if _, err := logFile.WriteString(fmt.Sprintf("%v\n", messageData)); err != nil {
			return err
		}
		if err := logFile.Close(); err != nil {
			return err
		}
	} else {
		logFile, err := os.OpenFile(fmt.Sprintf("log/%s.txt", room), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		messageData := fmt.Sprintf(`{"from":"%s", "type":"%s", "message":"%s", "room":"%s", "to":"%s"}`, from, MESSAGE_CHAT, message, room, to)
		if _, err := logFile.WriteString(fmt.Sprintf("%v\n", messageData)); err != nil {
			return err
		}
		if err := logFile.Close(); err != nil {
			return err
		}
	}
	return nil
}

func readLog(nameFile string) (string, error) {
	file, err := os.Open(fmt.Sprintf("log/%s.txt", nameFile))
	if err != nil {
		return "", err
	}
	defer file.Close()

	temp := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		temp = append(temp, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}
	return fmt.Sprintf("[%v]", strings.Join(temp, ",")), nil
}
