package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/fabarj4/goroomchat"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content, err := ioutil.ReadFile("index.html")
		if err != nil {
			http.Error(w, "Could not open requested file", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%s", content)
	})

	//fungsi ini digunakan untuk melakukan chat
	http.HandleFunc("/ws", goroomchat.HandleWS)

	//fungsi ini untuk menampilkan semua room yang sedang digunakan
	http.HandleFunc("/rooms/", goroomchat.HandleRooms)

	fmt.Println("Server starting at :8080")
	http.ListenAndServe(":8080", nil)
}
