package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"
	"vk_friends/logger"
)

var (
	tmpl         = template.Must(template.ParseGlob("templates/*.html"))
	clientID     = "51678013"
	clientSecret = "CCa6iiJfty8WVDEPtapJ"
	redirectURI  = "http://localhost:8080/me"
	state        = "12345"
)


func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/me", me)
	logger.Info.Println("-> Server has started")
	logger.Info.Println(http.ListenAndServe(":8080", nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("https://oauth.vk.com/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s", clientID, redirectURI, state)
	err := tmpl.ExecuteTemplate(w, "index.html", url)
	if err != nil {
		logger.Ferror(err)
	}
}

func me(w http.ResponseWriter, r *http.Request) {
	client := http.Client{Timeout: 5 * time.Second}

	stateTemp := r.URL.Query().Get("state")
	if stateTemp != state {
		logger.Error.Fatalln("state query param do not match original one, got=", stateTemp)
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		logger.Error.Fatalln("code query param is not provided")
	}

	// Получение токена
	url := fmt.Sprintf("https://oauth.vk.com/access_token?grant_type=authorization_code&code=%s&redirect_uri=%s&client_id=%s&client_secret=%s",
		code, redirectURI, clientID, clientSecret)
	
	req, _ := http.NewRequest("POST", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		logger.Ferror(err)
	}
	defer resp.Body.Close()
	token := struct {
		AccessToken string `json:"access_token"`
	}{}
	bytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bytes, &token)

	// Получение данных
	url = fmt.Sprintf("https://api.vk.com/method/%s?fields=photo_50&v=5.131&access_token=%s", "friends.get", token.AccessToken)
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Ferror(err)
	}
	resp, err = client.Do(req)
	if err != nil {
		logger.Ferror(err)
	}
	defer resp.Body.Close()
	bytes, err = io.ReadAll(resp.Body)
	if err != nil {
		logger.Ferror(err)
	}

	// Обработка и вывод списка друзей
	response := ResponseFriends{}
	jErr := json.Unmarshal(bytes, &response)
	if jErr != nil {
		logger.Ferror(jErr)
	}

	err = tmpl.ExecuteTemplate(w, "me.html", response.Friends.Persons)
	if err != nil {
		logger.Ferror(err)
	}
}

type ResponseFriends struct {
	Friends Friends `json:"response"`
}

type Friends struct {
	Count 	int 		`json:"count"`
	Persons []Person 	`json:"items"`
}

type Person struct {
	Id uint32 `json:"id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
}
