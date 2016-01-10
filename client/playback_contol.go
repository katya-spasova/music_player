package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

const HOST = "http://localhost:8765/"

func main() {
	getAlive()
}

func getAlive() {
	response, err := http.Get(HOST)
	if err != nil {
		fmt.Println("The service is not alive")
		return
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	text := string(body[:])
	fmt.Println(text)
}
