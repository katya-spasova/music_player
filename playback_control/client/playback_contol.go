// Package client provides the primitives for a command-line client for music_player
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

// Client struct holds the host on which music_player is running
type Client struct {
	Host string
}

// getAlive checks if music_player is running
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

// PerformAction uses the entered action and name to construct HTTP request and send it to music_player
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

// isLocalhostCall checks if music_player's host is localhost
func (client *Client) isLocalhostCall() bool {
	return strings.HasPrefix(client.Host, "http://localhost") ||
		strings.HasPrefix(client.Host, "https://localhost") ||
		strings.HasPrefix(client.Host, "http://127.") ||
		strings.HasPrefix(client.Host, "https://127.")
}

// determineHttpMethod determines which method (GET, POST or PUT) is going to be used for the
// HTTP request
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

// formUrl uses the entered action and name to construct the URL that is going to call music_player
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

// ResponseContainer struct is used to hold the unmarshalled json response of music_player
// Contains code (0 for succes, 1 for failure), message and a list of file names
type ResponseContainer struct {
	Code    int
	Message string
	Data    []string
}

// performCall send HTTP request to music_player, gets json the response and unmarshals it
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

// getDisplayMessage creates a string message based on the response of music_player
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

// escape does the escaping of query string
func escape(urlPath string) string {
	return strings.Replace(url.QueryEscape(urlPath), "+", "%20", -1)
}
