package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

const host = "http://localhost:8765/"

func GetAlive() string {
	response, err := http.Get(host)
	if err != nil {
		fmt.Println("The service is not alive")
		return ""
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	text := string(body[:])
	fmt.Println(text)
	return text
}
