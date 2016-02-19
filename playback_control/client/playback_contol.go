package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Client struct {
	Host string
}

func (client *Client) getAlive() string {
	response, err := http.Get(client.Host)
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
	path := name
	if action != "save" && client.isLocalhostCall() {
		var err error = nil
		path, err = filepath.Abs(name)
		if err != nil {
			//ignore error, try with name
			path = name
		}
		_, err = os.Stat(path)
		if os.IsNotExist(err) {
			// this can be a saved playlist, so try with name
			path = name
		}
	}
	response, err := performCall(determineHttpMethod(action), client.formUrl(action, path))
	return getDisplayMessage(response, err)
}

func (client *Client) isLocalhostCall() bool {
	return strings.HasPrefix(client.Host, "http://localhost") ||
		strings.HasPrefix(client.Host, "https://localhost") ||
		strings.HasPrefix(client.Host, "http://127.") ||
		strings.HasPrefix(client.Host, "https://127.")
}

func determineHttpMethod(action string) (method string) {
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
		requestUrl = client.Host + action

	case "add",
		"play",
		"save":
		requestUrl = client.Host + action + "/" + escape(name)
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

	err = json.Unmarshal(body, &container)

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
				buffer.WriteString("\n")
				buffer.WriteString(element)
			}
		}
		return buffer.String()
	}
}

func escape(urlPath string) string {
	return strings.Replace(url.QueryEscape(urlPath), "+", "%20", -1)
}
