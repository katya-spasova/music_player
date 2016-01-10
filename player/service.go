package main

import (
	"fmt"
	"net/http"

	"goji.io"
	"goji.io/pat"
	"golang.org/x/net/context"
)

// Shows if the service is alive
func alive(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "I'm alive")
}

// Starts playing a file, files from directory or a playlist immediately - current queue is cleared
// The result json contains the names of the files to be played
// or error message if song is not found
func play(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	name := pat.Param(ctx, "name")
	// todo: play file instead and return json
	fmt.Fprintf(w, "Playing, %s!", name)
}

// Pauses the current song
// The result json contains the filename of the paused song
// or error message if no song is playing at the moment
func pause(w http.ResponseWriter, r *http.Request) {

}

// Resumes paused song
// The result json contains the filename of the running song
// or error message if no song was paused
func resume(w http.ResponseWriter, r *http.Request) {

}

// Stops the playback - playing queue is cleared i.e. playback cannot be resumed
// The result json contains message that playback is stopped
func stop(w http.ResponseWriter, r *http.Request) {

}

// Skips playing the current song and starts playing the next from the queue
// The result json contains the filename of the running song
// or error message if there is no next song
func next(w http.ResponseWriter, r *http.Request) {

}

// Skips playing the current song and starts playing the previous from the queue
// The result json contains the filename of the running song
// or error message if there is no previous song
func previous(w http.ResponseWriter, r *http.Request) {

}

// Gets the filename of the current song
// The result json contains the filename of the running song
// or error message if no song is playing at the moment
func getCurrentSongInfo(w http.ResponseWriter, r *http.Request) {

}

// Add a song, directory or playlist to the play queue - songs will be played after all others in the queue
// Will play the song if there is no songs in the queue
// The result json contains the filename of the added song, directory or playlist
// or error message if song is not found
func addToQueue(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	filename := pat.Param(ctx, "name")
	// todo: add to queue instead and return json
	fmt.Fprintf(w, "Adding , %s!", filename)
}

// Saves the current queue as a playlist
// The result json contains the filename if the playlist
// or error message if queue is empty
func saveAsPlaylist(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	name := pat.Param(ctx, "name")
	// todo: save the playlist instead and return json
	fmt.Fprintf(w, "Saving playlist , %s!", name)
}

// List the saved playlists
// The result json contains the filenames of the already saved playlists
// or error message of no playlists exit
func listPlaylists(w http.ResponseWriter, r *http.Request) {

}

var player Player

func main() {
	// init the player
	player = Player{}
	player.init()

	// clean up
	defer player.clear()

	// service handle functions
	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/"), alive)
	mux.HandleFuncC(pat.Put("/play/:name"), play)
	mux.HandleFunc(pat.Post("/pause"), pause)
	mux.HandleFunc(pat.Post("/resume"), resume)
	mux.HandleFunc(pat.Put("/stop"), stop)
	mux.HandleFunc(pat.Post("/next"), next)
	mux.HandleFunc(pat.Post("/previous"), previous)
	mux.HandleFunc(pat.Get("/songinfo"), getCurrentSongInfo)
	mux.HandleFuncC(pat.Post("/add/:name"), addToQueue)
	mux.HandleFuncC(pat.Put("/save/:name"), saveAsPlaylist)
	mux.HandleFunc(pat.Get("/playlists"), listPlaylists)

	// start the service
	http.ListenAndServe(":8765", mux)
}
