package main

import (
	"errors"
	"html/template"
	"net/http"
	"os/exec"
	"runtime"
	"vk_friends/logger"
)

var (
	tmpl         = template.Must(template.ParseGlob("templates/*.html"))
)

const (
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

func openUrl(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = errors.New("unsupported platform")
	}

	if err != nil {
		logger.Error.Println(err)
	}
}
