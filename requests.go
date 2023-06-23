package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
	"vk_friends/logger"
)

func index(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("https://oauth.vk.com/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s", clientID, redirectURI, state)
	err := tmpl.ExecuteTemplate(w, "index.html", url)
	if err != nil {
		logger.Ferror(err)
	}
}

func me(w http.ResponseWriter, r *http.Request) {
	client := http.Client{Timeout: 5 * time.Second}

	// Получение кода
	code, err := getAuthCode(r)
	if err != nil {
		logger.Error.Println(err)
		return
	}

	// Получение токена
	token, err := getToken(code, client)
	if err != nil {
		logger.Error.Println(err)
		return
	}

	// Получение данных
	bytes, err := getData(token, client)
	if err != nil {
		logger.Error.Fatalln(err)
	}

	// Обработка и вывод списка друзей
	data, err := dataAnalysis(bytes)
	if err != nil {
		logger.Error.Fatalln(err)
	}

	// Вывод изменений в списке друзей
	err = tmpl.ExecuteTemplate(w, "me.html", data.Changes)
	if err != nil {
		logger.Error.Fatalln(err)
	}

}

func getAuthCode(r *http.Request) (code string, err error) {
	stateTemp := r.URL.Query().Get("state")

	if stateTemp != state {
		err = errors.New("state query param do not match original one, got=" + stateTemp)
		return "", err
	}

	code = r.URL.Query().Get("code")
	if code == "" {
		err = errors.New("code query param is not provided")
	}
	return code, err
}

func getToken(code string, client http.Client) (token string, err error) {
	url := fmt.Sprintf("https://oauth.vk.com/access_token?grant_type=authorization_code&code=%s&redirect_uri=%s&client_id=%s&client_secret=%s",
		code, redirectURI, clientID, clientSecret)
	
	req, _ := http.NewRequest("POST", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	tok := struct {
		AccessToken string `json:"access_token"`
	}{}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(bytes, &tok)
	if err != nil {
		return "", err
	}
	return tok.AccessToken, err
}

func getData(token string, client http.Client) (bytes []byte, err error) {
	url := fmt.Sprintf("https://api.vk.com/method/%s?fields=photo_50&v=5.131&access_token=%s", "friends.get", token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes, err
}