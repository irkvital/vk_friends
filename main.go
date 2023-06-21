package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
	"vk_friends/logger"
)

var (
	tmpl         = template.Must(template.ParseGlob("templates/*.html"))
	clientID     = "51678013"
	clientSecret = "CCa6iiJfty8WVDEPtapJ"
	localhost    = "http://localhost:8080"
	redirectURI  = localhost + "/me"
	state        = "12345"
)


func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/me", me)
	openUrl(localhost)
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
		logger.Error.Println("state query param do not match original one, got=", stateTemp)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		logger.Error.Println("code query param is not provided")
		return
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

	data := startFriendList(&response)

	err = tmpl.ExecuteTemplate(w, "me.html", data.Changes)
	if err != nil {
		logger.Ferror(err)
	}

}

type ChangeFriends struct {
	Persons []Person
	Changes []Change
}

type ResponseFriends struct {
	Friends Friends `json:"response"`
}

type Friends struct {
	Count 	int 		`json:"count"`
	Persons []Person 	`json:"items"`
}

type Person struct {
	Id 			uint32 	`json:"id"`
	FirstName 	string 	`json:"first_name"`
	LastName 	string 	`json:"last_name"`
}

type Change struct {
	Person 	Person
	Data 	string
	Status 	string
}



func startFriendList(r *ResponseFriends) ChangeFriends {
	filename := "data.txt"
	data := ChangeFriends{}


	byte, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		fileCreate(filename)
		byte, err = os.ReadFile(filename)
		if err != nil {
			logger.Error.Fatalln(err)
		}
	} else if err != nil {
		logger.Error.Fatalln(err)
	}


	if len(byte) != 0 {
		jErr := json.Unmarshal(byte, &data)
		if jErr != nil {
			logger.Error.Fatalln("Can't read correct data from file", filename)
		}
		// logger.Debug.Println(data)
		// Сравнение данных
		compareData(r, &data)
	} else {
		// Если файл с данными пустой, записывается текущий список друзей
		data.Persons = r.Friends.Persons
	}
	WriteFile(filename, &data)
	return data
}

func compareData(r *ResponseFriends, f *ChangeFriends) {
	// Карта с id текущих друзей
	respData := make(map[uint32]Change)
	for _, val := range r.Friends.Persons {
		respData[val.Id] = Change{val, "", "added"}
	}
	// Сравнение старого списка друзей с новым
	for i := range f.Persons {
		key := f.Persons[i].Id
		_ , found := respData[key]
		if found {
			delete(respData, key)
		} else {
			respData[key] = Change{f.Persons[i], "", "deleted"}
		}
	}
	// Добавление изменений
	for _, val := range respData {
		val.Data = time.Now().Format("01.02.2006")
		f.Changes = append(f.Changes, val)
	}
	// Перезаписывание списка друзей
	f.Persons = r.Friends.Persons
}



func fileCreate(filename string) {
	_, fileErr := os.Create(filename)
	if fileErr != nil {
		logger.Error.Fatalln(fileErr)
	}
	logger.Info.Println("File", filename, "created")
}

func WriteFile(filename string, r *ChangeFriends) {
	// Запись данных
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		logger.Error.Fatalln(err)
	}
	defer file.Close()
	byteData, _ := json.MarshalIndent(r, "", "")
	count, _ := file.Write(byteData)
	logger.Info.Println(count, "bytes was written to file", filename)
}

func openUrl(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		logger.Error.Println(err)
	}
}