package player

import (
	"fmt"
	"net/http"

	"encoding/json"
	"github.com/krig/go-sox"
	"goji.io"
	"goji.io/pat"
	"golang.org/x/net/context"
	"strings"
	"sync"
)

// Success and error codes
const (
	success = iota
	failure
)

// Messages
const no_sox_in_msg string = "SoX failed to open input file"
const no_sox_out_msg string = "Sox failed to open output device"
const file_not_found_msg string = "File cannot be found"
const playlist_not_found_msg string = "Playlist cannot be found"
const no_playlists_msg string = "Currently there are no saved playlists"
const format_not_supported_msg string = "Format is not supported"
const cannot_pause_msg string = "Cannot pause. No song is playing"
const cannot_resume_msg = "Cannot resume. No song was paused"
const cannot_next_msg = "Cannot play next song. No next song in queue"
const cannot_previous_msg = "Cannot play previous song. No previous song in queue"
const cannot_get_info_msg = "There is no current song in the queue"
const cannot_save_playlist_msg = "Cannot save playlist"
const cannot_save_empty_queue_msg = "Queue is empty and cannot be saved as playlist"
const cannot_get_queue_info_msg = "Cannot get queue info. Queue is empty"

const started_playing_info = "Started playing"
const added_to_queue_info = "Added to queue"
const paused_song_info = "Song is paused"
const resume_song_info = "Song is resumed"
const playback_stopped_info = "Playback is stopped and cleaned"
const current_song_info = "The filename of the current song"
const current_queue_info = "The filenames in the current queue"
const queue_saved_as_playlist = "The queue is saved as a playlist"
const playlists_info = "A list of all saved playlists"
const queue_info = "Queue content"

// ResponseContainer defines the format of the web service's response
// It contains code - 0 for success and 1 for error, message that explains actions is performed
// and data which is a list of file names
type ResponseContainer struct {
	// 0 for success, 1 for failure
	Code int
	// Error message or Info message
	Message string
	// Filename (list if filenames)
	Data []string `json:"Data,omitempty"`
}

// writeHttpResponse writes response
func writeHttpResponse(w http.ResponseWriter, container ResponseContainer) {
	message, err1 := json.Marshal(container)
	if err1 != nil {
		http.Error(w, err1.Error(), http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		if message != nil {
			w.Write(message)
		}
	}
}

// getResponseContainer wraps data and error in ResponseContainer
func getResponseContainer(data []string, err error) ResponseContainer {
	container := ResponseContainer{}
	if err != nil {
		container.Code = failure
		container.Message = err.Error()
	} else {
		container.Code = success
		container.Data = data
	}

	return container
}

// alive shows if the service is alive
func alive(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "I'm alive")
}

// playerToServiceResponse constructs the response in the format of the service and writes it
func playerToServiceResponse(w http.ResponseWriter, data []string, err error, successMessage string) {
	container := getResponseContainer(filterPath(data), err)
	if err == nil {
		container.Message = successMessage
	}
	writeHttpResponse(w, container)
}

// filterPath removes the path from list of filenames. Only the part after the last slash remains
func filterPath(data []string) []string {
	filtered := make([]string, 0, len(data))
	for _, element := range data {
		lastSlashIndex := strings.LastIndex(element, "/")
		filtered = append(filtered, element[lastSlashIndex+1:])
	}
	return filtered
}

// Starts playing a file, files from directory or a playlist immediately - current queue is cleared
// The result json contains the names of the files to be played
// or error message if song is not found, format is unsupported or SoX cannot play the file
func play(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	name := pat.Param(ctx, "name")
	data, err := player.play(name)
	playerToServiceResponse(w, data, err, started_playing_info)
}

// Pauses the current song
// The result json contains the filename of the paused song
// or error message if no song is playing at the moment
func pause(w http.ResponseWriter, r *http.Request) {
	data, err := player.pause()
	playerToServiceResponse(w, []string{data}, err, paused_song_info)
}

// Resumes paused song
// The result json contains the filename of the running song
// or error message if no song was paused
func resume(w http.ResponseWriter, r *http.Request) {
	data, err := player.resume()
	playerToServiceResponse(w, []string{data}, err, resume_song_info)
}

// Stops the playback - playing queue is cleared i.e. playback cannot be resumed
// The result json contains message that playback is stopped
func stop(w http.ResponseWriter, r *http.Request) {
	player.stop()
	playerToServiceResponse(w, []string{}, nil, playback_stopped_info)
}

// Starts playing the next from the queue
// Skips playing the current song (if not paused)
// The result json contains the filename of the running song
// or error message if there is no next song
func next(w http.ResponseWriter, r *http.Request) {
	data, err := player.next()
	playerToServiceResponse(w, []string{data}, err, started_playing_info)
}

// Starts playing the previous from the queue
// Skips playing the current song (if not paused)
// The result json contains the filename of the running song
// or error message if there is no previous song
func previous(w http.ResponseWriter, r *http.Request) {
	data, err := player.previous()
	playerToServiceResponse(w, []string{data}, err, started_playing_info)
}

// Gets the filename of the current song
// The result json contains the filename of the running song
// or error message if no song is playing at the moment
func getCurrentSongInfo(w http.ResponseWriter, r *http.Request) {
	data, err := player.getCurrentSongInfo()
	playerToServiceResponse(w, []string{data}, err, current_song_info)
}

// Add a song, directory or playlist to the play queue - songs will be played after all others in the queue
// Will play the song if there is no songs in the queue
// The result json contains the filename of the added song, directory or playlist
// or error message if song is not found, format is unsupported or Sox cannot play the file
func addToQueue(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	filename := pat.Param(ctx, "name")
	data, err := player.addToQueue(filename)
	playerToServiceResponse(w, data, err, added_to_queue_info)
}

// saveAsPlaylist saves the current queue as a playlist
// The result json contains the filename of the playlist
// or error message if queue is empty or playlist cannot be saved
func saveAsPlaylist(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	name := pat.Param(ctx, "name")
	data, err := player.saveAsPlaylist(name)
	playerToServiceResponse(w, []string{data}, err, queue_saved_as_playlist)
}

// listPlaylists list the saved playlists
// The result json contains the filenames of the already saved playlists
// or error message of no playlists exit
func listPlaylists(w http.ResponseWriter, r *http.Request) {
	data, err := player.listPlaylists()
	playerToServiceResponse(w, data, err, playlists_info)
}

// getQueueInfo Displays all songs in the queue
// The result json contains all filenames in the current queue
// or an error message if queue is empty
func getQueueInfo(w http.ResponseWriter, r *http.Request) {
	data, err := player.getQueueInfo()
	playerToServiceResponse(w, data, err, queue_info)
}

var player musicPlayer

//InitService creates a mux and initializes handle functions for music_player
func InitService() *goji.Mux {
	player = musicPlayer{playQueueMutex: &sync.Mutex{}}
	player.init()

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
	mux.HandleFunc(pat.Get("/queueinfo"), getQueueInfo)

	return mux
}

// WaitEnd is used to wait the end of the playing the songs in the play queue
func WaitEnd() {
	player.waitEnd()
}

// Start starts the music_player web service
func Start() {
	// init the player
	mux := InitService()
	// init sox
	if !sox.Init() {
		fmt.Errorf("sox is not found")
	}
	// clean up
	defer sox.Quit()
	// start the service
	http.ListenAndServe(":8765", mux)
}
