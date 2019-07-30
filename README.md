
# Go ROOM CHAT
library ini adalah hasil pengembangan dari chat app webscoket noval agung, untuk dokumentasi asli  [klik disini](https://dasarpemrogramangolang.novalagung.com/C-28-golang-web-socket.html). dalam library ini semua chat akan disimpan dalam log dengan nama chat.txt. user dapat berkomunikasi dengan cara private chat atau group.

# Features
 - setiap user yang menggunakan chat menggunakan username
 - percakapan dilakukan dengan cara private chat atau group chat
 - chat tersimpan ke dalam log, log ini digunakan untuk mengambil chat yang sebelumnya ketika halaman tertutup atau melakukan refresh

# INSTALASI
1. clone repository ini
2. ambil gorilla/websocket
    `go get github.com/gorilla/websocket`
3. ambil gubrak dari noval agung
    `go get -u github.com/novalagung/gubrak`
# Menjalankan Contoh
masuk ke folder example lalu ketikan "go run example.go" .disini host yang digunakan localhost dengan port 8080 hal ini bisa disesuaikan
untuk pengaksesan chat :
 - private chat
    localhost:8080?to= (nama user) misal : localhost:8080?to=admin
 - group chat
    localhost:8080/(nama group) misal : localhost:8080/grupa/
 - untuk melihat group yang aktif dan tersedia
   localhost:8080/rooms/

# Thanks
[Noval Agung](https://github.com/novalagung/dasarpemrogramangolang) - tutorial web socket


# Contact
jika terdapat pertanyaan dapat email dengan subjek goroomchat ke fabarj4@gmail.com
