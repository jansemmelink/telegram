package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/gorilla/pat"
	telegram "github.com/jansemmelink/telegram/api"
)

func main() {
	//e.g. botToken = "1516746894:AAErH_tiDWX5Jjr-v1wHiMUzFc0vs7PKiE8"
	telegramURLPtr := flag.String("u", "https://api.telegram.org", "Telegram URL")
	botTokenPtr := flag.String("t", "", "Telegram Bot Token")
	webHookPtr := flag.String("w", "https://localhost:8443/bot", "Web hook ULR must be HTTPS")
	flag.Parse()
	if *botTokenPtr == "" {
		panic("-t <token> not specified")
	}
	http.ListenAndServe("localhost:12345", newApp(*telegramURLPtr, *botTokenPtr, *webHookPtr))
}

func newApp(telegramURL, botToken string, webHook string) http.Handler {
	r := pat.New()
	myApp := app{
		Handler:     r,
		telegramURL: telegramURL,
		botToken:    botToken,
		webHook:     webHook,
	}
	r.Post("/update", myApp.handleUpdateFromBot)
	r.Post("/", myApp.postHandler)
	r.Get("/", myApp.unknownHandler)

	//todo: verify token
	//e.g. curl -XGET 'https://api.telegram.org/bot1609917215:AAEgCizNgEQzFgNz45BeK8rnmrWdqOXsGvA/getMe'
	//{"ok":true,"result":{"id":1609917215,"is_bot":true,"first_name":"ShopFlow","username":"ShopFlowBot","can_join_groups":true,"can_read_all_group_messages":false,"supports_inline_queries":false}}
	// {
	// 	"ok": true,
	// 	"result": {
	// 	  "id": 1609917215,
	// 	  "is_bot": true,
	// 	  "first_name": "ShopFlow",
	// 	  "username": "ShopFlowBot",
	// 	  "can_join_groups": true,
	// 	  "can_read_all_group_messages": false,
	// 	  "supports_inline_queries": false
	// 	}
	// }
	var getMeRes telegram.GetMeResponse
	err := myApp.do("/getMe", nil, &getMeRes)
	if err != nil {
		panic(err)
	}
	fmt.Printf("getMe -> %+v\n", getMeRes)

	//set webHook for async updates from the bot
	var setWebHookResponse telegram.SetWebHookResponse
	err = myApp.do("/setWebHook", telegram.SetWebHookRequest{URL: myApp.webHook}, &setWebHookResponse)
	if err != nil {
		panic(err)
	}
	fmt.Printf("setWebHook -> %+v\n", setWebHookResponse)

	return myApp
}

type app struct {
	http.Handler
	telegramURL string //"https://api.telegram.org"
	botToken    string
	webHook     string //https URL in nginx that proxy to this server
}

func (app) unknownHandler(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "unknown", http.StatusNotFound)
}

func (a app) postHandler(httpRes http.ResponseWriter, httpReq *http.Request) {
	fmt.Printf("HTTP %s %s\n", httpReq.Method, httpReq.URL.Path)
	var reqBody interface{}
	if httpReq.Header.Get("Content-Type") == "application/json" {
		//parse JSON body
		if err := json.NewDecoder(httpReq.Body).Decode(&reqBody); err != nil {
			http.Error(httpRes, "invalid json: "+err.Error(), http.StatusBadRequest)
		}
	} else {
		fmt.Printf("Not reading content-type: \"%s\"\n", httpReq.Header.Get("Content-Type"))
	}
	fmt.Printf("Request body: %+v\n", reqBody)

	//call the bot
	var botRes interface{}
	err := a.do(httpReq.URL.Path, reqBody, &botRes)
	if err != nil {
		http.Error(httpRes, fmt.Sprintf("bot failed: %+v", err), http.StatusInternalServerError)
	}

	//respond with JSON data
	jsonResponse, _ := json.Marshal(botRes)
	httpRes.Header().Set("Content-Type", "application/json")
	httpRes.Write(jsonResponse)
	httpRes.Write([]byte("\n"))
}

func (a app) handleUpdateFromBot(httpRes http.ResponseWriter, httpReq *http.Request) {
	fmt.Printf("HTTP %s %s\n", httpReq.Method, httpReq.URL.Path)
	var reqBody interface{}
	if httpReq.Header.Get("Content-Type") == "application/json" {
		//parse JSON body
		if err := json.NewDecoder(httpReq.Body).Decode(&reqBody); err != nil {
			http.Error(httpRes, "invalid json: "+err.Error(), http.StatusBadRequest)
		}
		fmt.Printf("Request Body: %+v\n", reqBody)
	}

	http.Error(httpRes, "not yet implemented", http.StatusInternalServerError)
}

func (a app) do(path string, reqData interface{}, resDataPtr interface{}) error {
	url := a.telegramURL + "/bot" + a.botToken
	url += path
	var httpRes *http.Response
	if reqData == nil {
		fmt.Printf("HTTP GET %s\n", url)
		var err error
		httpRes, err = http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to get from bot: %+v", err)
		}
	} else {
		fmt.Printf("HTTP POST %s\n", url)
		var err error
		jsonBody, _ := json.Marshal(reqData)
		httpRes, err = http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("failed to post to bot: %+v", err)
		}
	}

	// The response contains a JSON object, which always has a Boolean field 'ok'
	// and may have an optional String field 'description' with a human-readable description of the result.
	// If 'ok' equals true, the request was successful and the result of the query can be found in the 'result' field.
	// In case of an unsuccessful request, 'ok' equals false and the error is explained in the 'description'.
	// An Integer 'error_code' field is also returned, but its contents are subject to change in the future.
	// Some errors may also have an optional field 'parameters' of the type ResponseParameters,
	// which can help to automatically handle the error.
	// All methods in the Bot API are case-insensitive.
	// All queries must be made using UTF-8.
	switch httpRes.Header.Get("Content-Type") {
	case "application/json":
		var res telegram.Response
		if err := json.NewDecoder(httpRes.Body).Decode(&res); err != nil {
			return fmt.Errorf("failed to decode JSON response: %+v", err)
		}
		if !res.OK {
			return fmt.Errorf("failed: %v", res.Desription)
		}

		//encode and decode the result
		jsonResult, _ := json.Marshal(res.Result)
		if err := json.Unmarshal(jsonResult, resDataPtr); err != nil {
			return fmt.Errorf("failed to parse result as %T: %v", resDataPtr, err)
		}
		return nil
	default:
		return fmt.Errorf("unknown Content-Type: %s", httpRes.Header.Get("Content-Type"))
	}
}
