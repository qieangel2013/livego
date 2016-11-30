package main

import (
	//"bufio"
	//"./lib/client"
	"./lib/websocket"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/big"
	"net/http"
	"runtime"
)

type userinfo struct {
	name string
	img  string
}
type Client struct {
	id   string
	conn *websocket.Conn
	userinfo
}

type message struct {
	Data  string
	Mtype string
	Img   string
}

var member = make(map[string]*Client)

func getclient(ws *websocket.Conn) bool {
	for _, v := range member {
		if v.conn == ws {
			return true
		}
	}
	return false
}

func getnun() string {
	rnd, _ := rand.Int(rand.Reader, big.NewInt(12))
	num := fmt.Sprintf("%v", rnd)
	return num
}

func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func guid() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return GetMd5String(base64.URLEncoding.EncodeToString(b))
}

func (m *Client) addclient(ws *websocket.Conn) *Client {
	m.conn = ws
	return m
}

var i int = 0

func pwint(ws *websocket.Conn) {
	var err error
	for {

		var reply string

		if err = websocket.Message.Receive(ws, &reply); err != nil {
			fmt.Println("Can't receive")
			break
		}
		ok := getclient(ws)
		if ok {
			goto sendmsg
		} else {
			uid := guid()
			i = i + 1
			var mymes message
			json.Unmarshal([]byte(reply), &mymes)
			user := userinfo{fmt.Sprintf("游客%d说：", i), fmt.Sprintf("/public/images/%s.jpg", getnun())}
			client := Client{uid, ws, user}
			client.addclient(ws)
			member[uid] = &client
			/*mymes.Data = fmt.Sprintf("%s%s", user.name, mymes.Data)
			mymes.Img = user.img
			msg, _ := json.Marshal(mymes)
			fmt.Println(string(msg))
			if err = websocket.Message.Send(ws, string(msg)); err != nil {
				ok := getclient(ws)
				if ok {
					delete(member, client.id)
				}
				break
			}*/
		}
		//fmt.Println("Received back from client: " + reply)
	sendmsg:
		for k, v := range member {
			if v.conn != ws {
				var mymes message
				json.Unmarshal([]byte(reply), &mymes)
				mymes.Data = fmt.Sprintf("%s%s", v.userinfo.name, mymes.Data)
				mymes.Img = v.userinfo.img
				msg, _ := json.Marshal(mymes)
				if err = websocket.Message.Send(v.conn, string(msg)); err != nil {
					if member[k] != nil {
						delete(member, k)
					}
					break
				}
			}
		}

	}
}

func camera(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("./../public/views/camera.html")
		t.Execute(w, nil)
	} else {

	}
}
func live(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("./../public/views/live.html")
		t.Execute(w, nil)
	} else {

	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)
	fmt.Printf("LiveGoServer is ready...\n")
	http.Handle("/chat", websocket.Handler(pwint))
	http.Handle("/", http.FileServer(http.Dir("./../")))
	http.HandleFunc("/live", live)
	http.HandleFunc("/camera", camera)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("LiveGoServer:", err)
	}

}
