package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/gorilla/pat"
)

func main() {
	//e.g. botToken = "1516746894:AAErH_tiDWX5Jjr-v1wHiMUzFc0vs7PKiE8"
	botTokenPtr := flag.String("t", "", "Telegram Bot Token")
	flag.Parse()
	if *botTokenPtr == "" {
		panic("-t <token> not specified")
	}
	http.ListenAndServe("localhost:12345", newApp(*botTokenPtr))
}

func newApp(botToken string) http.Handler {
	r := pat.New()
	myApp := app{
		Handler:  r,
		botToken: botToken,
	}
	r.Post("/", myApp.postHandler)
	r.Get("/", myApp.unknownHandler)
	return myApp
}

type app struct {
	http.Handler
	botToken string
}

func (app) unknownHandler(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "unknown", http.StatusNotFound)
}

func (a app) postHandler(httpRes http.ResponseWriter, httpReq *http.Request) {
	var reqBody interface{}
	if httpReq.Header.Get("Content-Type") == "application/json" {
		//parse JSON body
		if err := json.NewDecoder(httpReq.Body).Decode(&reqBody); err != nil {
			http.Error(httpRes, "invalid json: "+err.Error(), http.StatusBadRequest)
		}
	}

	//call the bot
	botRes, err := a.do(httpReq.URL.Path, reqBody)
	if err != nil {
		http.Error(httpRes, fmt.Sprintf("bot failed: %+v", err), http.StatusInternalServerError)
	}

	//respond with JSON data
	jsonResponse, _ := json.Marshal(botRes)
	httpRes.Header().Set("Content-Type", "application/json")
	httpRes.Write(jsonResponse)
	httpRes.Write([]byte("\n"))
}

func (a app) do(path string, data interface{}) (interface{}, error) {
	url := "https://api.telegram.org/bot" + a.botToken
	url += path
	fmt.Printf("HTTP GET %s\n", url)
	httpRes, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to access bot: %+v", err)
	}
	switch httpRes.Header.Get("Content-Type") {
	case "application/json":
		var res interface{}
		if err := json.NewDecoder(httpRes.Body).Decode(&res); err != nil {
			return nil, fmt.Errorf("failed to decode JSON response: %+v", err)
		}
		return res, nil
	default:
		return nil, fmt.Errorf("unknown Content-Type: %s", httpRes.Header.Get("Content-Type"))
	}
}
