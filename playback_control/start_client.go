package main

import "github.com/katya-spasova/music_player/playback_control/client"
import (
	"flag"
	"fmt"
)

const host = "http://localhost:8765/"

func isValidAction(action string) bool {
	switch action {
	case
		"play",
		"stop",
		"pause",
		"resume",
		"next",
		"previous",
		"add",
		"songinfo",
		"queueinfo",
		"playlists",
		"save":
		return true
	}
	return false
}

func main() {
	action := flag.String("action", "stop",
		"Use one of: play/stop/pause/resume/next/previous/add/songinfo/queueinfo/playlists/save")
	if !isValidAction(*action) {
		fmt.Println(`Unknown action. Use one of: play/stop/pause/resume/next
		/previous/add/songinfo/queueinfo/playlists/save`)
	}

	name := flag.String("name", "", "Name of a song, a directory or a playlist")

	if (*action == "play" || *action == "add" || *action == "save") && len(*name) == 0 {
		fmt.Println("file, directory or playlist name is required with this action")
	}

	specifiedHost := flag.String("host", host, "Specify the host")

	cl := client.Client{host: *specifiedHost}
	fmt.Print(cl.PerformAction(*action, *name))
}
