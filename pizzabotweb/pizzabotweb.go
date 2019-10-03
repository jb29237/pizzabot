package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

//twilio account SID
var twilaccount = os.Getenv("TWILIOACCOUNT")

//twilio account authorization token
var twiltoken = os.Getenv("TWILIOTOKEN")

//tunnel url in format https://fj4hahd7.ngrok.io
var ngrokurl = os.Getenv("PIZZANGROKURL")

// sets the number you will be calling
var callto = os.Getenv("PIZZACALLTO")

//your twilio number that can make outbound calls
var callfrom = os.Getenv("PIZZACALLFROM")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {

	println(twilaccount)
	println(twiltoken)
	http.HandleFunc("/twiml", xmlpost)
	http.HandleFunc("/audiooutput", audioWebsocket)
	http.HandleFunc("/call", call)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download", downloadHandler)

	http.ListenAndServe(":3030", nil)

}

func call(w http.ResponseWriter, r *http.Request) {
	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + twilaccount + "/Calls.json"
	println(urlStr)
	// Build out the data for our message
	v := url.Values{}
	v.Set("To", callto)
	v.Set("From", callfrom)
	v.Set("Url", "https://"+ngrokurl+"/twiml")
	println(ngrokurl)
	rb := *strings.NewReader(v.Encode())

	// Create Client
	client := &http.Client{}

	req, _ := http.NewRequest("POST", urlStr, &rb)
	req.SetBasicAuth(twilaccount, twiltoken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// make request
	resp, _ := client.Do(req)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		err := json.Unmarshal(bodyBytes, &data)
		if err == nil {
			fmt.Println(data["sid"])
		}
	} else {
		fmt.Println(resp.Status)
	}
}

func xmlgen(x1 string, x2 string) []byte {
	var xmlurl = x1
	var xmlmp3 = x2
	type Response struct {
		XMLName xml.Name `xml:"Response"`
		Start   struct {
			XMLName xml.Name `xml:"Start"`
			Stream  struct {
				XMLName xml.Name `xml:"Stream"`
				URL     string   `xml:"url,attr"`
				Text    string   `xml:",chardata"`
			}
		}
		Play  string `xml:"Play"`
		Pause struct {
			XMLName xml.Name `xml:"Pause"`
			Length  string   `xml:"length,attr"`
		}
		Redirect struct {
			XMLName xml.Name `xml:"Redirect"`
			Method  string   `xml:"method,attr"`
			Text    string   `xml:",chardata"`
			//string  `xml:",charset"`
		}
	}

	twiml := &Response{Play: xmlmp3}
	twiml.Pause.Length = "5"
	twiml.Redirect.Method = "POST"
	twiml.Redirect.Text = xmlurl
	twiml.Start.Stream.URL = "wss://" + ngrokurl + "/audiooutput"
	sh, err := xml.Marshal(twiml)
	if err != nil {
		panic(err)
	}
	return sh
}

func xmlpost(w http.ResponseWriter, r *http.Request) {

	//twiml.Redirect.URL = "http://thisistheurl.com"
	v := xmlgen("https://"+ngrokurl+"/twiml", "https://"+ngrokurl+"/download")
	w.Header().Set("Content-Type", "application/xml")
	//w.Header().Add("Cache-Control:", "no-cache")
	//println(v)
	w.Write(v)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Create("./outputpost.mp3")
	if err != nil {
		panic(err)
	}
	n, err := io.Copy(file, r.Body)
	if err != nil {
		panic(err)
	}

	w.Write([]byte(fmt.Sprintf("%d bytes are recieved.\n", n)))
}
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, "outputpost.mp3")

}

func audioWebsocket(w http.ResponseWriter, r *http.Request) {

	//skip CORS till deploy to real world
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	// helpful log statement to show connections
	log.Println("Client Connected")

	reader(ws)
}

func reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// log that message for clarity
		f, err := os.OpenFile("./logfile", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		/*
				type AudioMessage struct {
					Media struct {
						//Track     string `json:"track"`
						//Chunk     string `json:"chunk"`
						//Timestamp string `json:"timestamp"`
						Payload string `json:"payload"`
					}
				}

			var maudio []byte
			json.Unmarshal(data, &maudio)
		*/
		defer f.Close()
		//fmt.Fprintf(f, "%s", data)
		dst := &bytes.Buffer{}
		json.Indent(dst, data, "", " ")
		fmt.Println(dst.String())
		if err := conn.WriteMessage(messageType, data); err != nil {
			log.Println(err)
			return
		}

	}
}
