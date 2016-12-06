package main

import (
	"./../config"
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
	"os"
)

type userinfo struct {
	name   string
	img    string
	isself bool
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

func getclient(ws *websocket.Conn) string {
	for k, v := range member {
		if v.conn == ws {
			return k
		}
	}
	return ""
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

var username string
var uzb string = ""

func pwint(ws *websocket.Conn) {
	defer func() {
		ws.Close()
	}()
	uid := guid()
	//logger.Println(i)
	if username == "" && uzb == "女主播" {
		username = uzb
	}
	user := userinfo{fmt.Sprintf("%s：", username), fmt.Sprintf("/public/images/%s.jpg", getnun()), false}
	username = ""
	client := Client{uid, ws, user}
	client.addclient(ws)
	member[uid] = &client
	for {
		var err error
		var reply string
		if err = websocket.Message.Receive(ws, &reply); err != nil {
			logger.Println("LiveGoServer:", err)
			break
		}
		for k, v := range member {
			if v.conn != ws {
				var mymes message
				json.Unmarshal([]byte(reply), &mymes)
				if mymes.Mtype == "mess" {
					mymes.Data = fmt.Sprintf("%s%s", member[getclient(ws)].userinfo.name, mymes.Data)
					mymes.Img = member[getclient(ws)].userinfo.img
				}
				msg, _ := json.Marshal(mymes)
				if err = websocket.Message.Send(v.conn, string(msg)); err != nil {
					delete(member, k)
					logger.Println("LiveGoServer:", err)
					break
				}
			} else {
				if !v.userinfo.isself {
					var mymesss message
					json.Unmarshal([]byte(reply), &mymesss)
					//if mymesss.Mtype == "mess" {
					mymesss.Mtype = "self"
					mymesss.Img = v.userinfo.img
					msg, _ := json.Marshal(mymesss)
					member[k].userinfo.isself = true
					if err = websocket.Message.Send(ws, string(msg)); err != nil {
						delete(member, k)
						logger.Println("LiveGoServer:", err)
						break
					}
					//}
				}
			}

		}
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "GET" {
		t, _ := template.ParseFiles("./views/index.html")
		t.Execute(w, nil)
	} else {
		_, ok := r.Form["name"]
		if ok {
			username = r.FormValue("name")
			http.Redirect(w, r, "/live", 301)
		} else {
			http.Redirect(w, r, "/index", 301)
		}

	}
}

func camera(w http.ResponseWriter, r *http.Request) {
	uzb = "女主播"
	if r.Method == "GET" {
		t, _ := template.ParseFiles("./views/camera.html")
		t.Execute(w, nil)
	} else {

	}
}
func live(w http.ResponseWriter, r *http.Request) {
	if username == "" {
		http.Redirect(w, r, "/index", 301)
	}
	if r.Method == "GET" {
		t, _ := template.ParseFiles("./views/live.html")
		t.Execute(w, nil)
	} else {

	}
}

var logfile, err = os.OpenFile(config.ServerLog, os.O_RDWR|os.O_CREATE, 0666)
var logger = log.New(logfile, "\r\n", log.Ldate|log.Ltime|log.Llongfile)

func main() {
	//fmt.Printf("LiveGoServer is ready...\n")
	go func() {
		http.Handle("/chat", websocket.Handler(pwint))
	}()
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	http.HandleFunc("/live", live)
	http.HandleFunc("/camera", camera)
	http.Handle("/", http.RedirectHandler("/index", 301))
	http.HandleFunc("/index", index)
	var config = config.ServerHost + ":" + config.ServerPort
	if err := http.ListenAndServe(config, nil); err != nil {
		logger.Println("LiveGoServer:", err)
		logfile.Close()
	}

}
