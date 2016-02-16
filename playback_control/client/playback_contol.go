package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Client struct {
	host string
}

func (client *Client) getAlive() string {
	response, err := http.Get(client.host)
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

func (client *Client) PerformAction(action string, name string) string {
	response, err := performCall(determineHttpMethod(action, name), client.formUrl(action, name))
	return getDisplayMessage(response, err)
}

func determineHttpMethod(action string, name string) (method string) {
	switch action {
	case
		"songinfo",
		"queueinfo",
		"playlists":
		method = "GET"

	case "next",
		"previous",
		"pause",
		"resume",
		"add":
		method = "POST"

	case "play",
		"save",
		"stop":
		method = "PUT"
	}
	return method
}

func (client *Client) formUrl(action string, name string) (requestUrl string) {
	switch action {
	case
		"songinfo",
		"queueinfo",
		"playlists",
		"next",
		"previous",
		"pause",
		"resume",
		"stop":
		requestUrl = client.host + action

	case "add",
		"play",
		"save":
		requestUrl = client.host + action + "/" + url.QueryEscape(name)
	}

	return requestUrl
}

//ResponseContainer - type used for success json response
type ResponseContainer struct {
	Code    int
	Message string
	Data    []string
}

func performCall(method string, url string) (ResponseContainer, error) {
	var res *http.Response
	var err error
	container := ResponseContainer{}

	if method == "GET" {
		res, err = http.Get(url)
	} else if method == "POST" {
		res, err = http.Post(url, "text/plain", nil)
	} else if method == "PUT" {
		client := &http.Client{}
		request, err1 := http.NewRequest("PUT", url, nil)
		if err1 != nil {
			return container, err1
		}
		res, err = client.Do(request)
	}

	if err != nil {
		return container, err
	}

	var body []byte
	body, err = ioutil.ReadAll(res.Body)
	res.Body.Close()

	if err != nil {
		return container, err
	}

	err = json.Unmarshal(body, container)

	if err != nil {
		return container, err
	}
	return container, nil
}

func getDisplayMessage(response ResponseContainer, err error) string {
	if err != nil {
		return err.Error()
	}

	if response.Code != 0 {
		return response.Message
	} else {
		buffer := bytes.NewBufferString(response.Message)
		data_list := response.Data
		if data_list != nil && len(data_list) > 0 {
			for _, element := range data_list {
				buffer.WriteString(" ")
				buffer.WriteString(element)
			}
		}
		return buffer.String()
	}
}
