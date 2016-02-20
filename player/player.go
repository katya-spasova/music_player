// Package player provides the implementation of music_player
package player

import "github.com/krig/go-sox"
import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// supported playlist type
const playlistsExtension = ".m3u"

// supportedExtensions are the file types music_player works with
var supportedExtensions []string = []string{
	"mp3",
	"ogg",
	"flac",
	"8svx",
	"aif",
	"aifc",
	"aiff",
	"aiffc",
	"al",
	"amb",
	"au",
	"avr",
	"cdda",
	"cdr",
	"cvs",
	"cvsd",
	"cvu",
	"dat",
	"dvms",
	"f32",
	"f4",
	"f64",
	"f8",
	"fssd",
	"gsm",
	"gsrt",
	"hcom",
	"htk",
	"ima",
	"ircam",
	"la",
	"lpc",
	"lpc10",
	"lu",
	"maud",
	"mp2",
	"nist",
	"prc",
	"raw",
	"s1",
	"s16",
	"s2",
	"s24",
	"s3",
	"s32",
	"s4",
	"s8",
	"sb",
	"sf",
	"sl",
	"sln",
	"smp",
	"snd",
	"sndr",
	"sndt",
	"sou",
	"sox",
	"sph",
	"sw",
	"txw",
	"u1",
	"u16",
	"u2",
	"u24",
	"u3",
	"u32",
	"u4",
	"u8",
	"ub",
	"ul",
	"uw",
	"vms",
	"voc",
	"vorbis",
	"vox",
	"wav",
	"wavpcm",
	"wve",
	"xa",
}

// musicPlayer struct represents the player. Holds player's state, playlist's directory and mutexes for synchronisation
type musicPlayer struct {
	sync.Mutex
	state          *state
	playQueueMutex *sync.Mutex
	playlistsDir   string
}

// State struct holds the state of the player i.e. chain of effects, playing status, playing start time of a song,
// player's song queue, current song
type state struct {
	chain          *sox.EffectsChain
	status         int
	startTime      time.Time
	durationPaused time.Duration
	queue          []string
	current        int
}

// player's possible statuses
const (
	playing = iota
	paused
	waiting
)

// init initialises player's state
func (player *musicPlayer) init() error {
	player.Lock()
	defer player.Unlock()
	player.state = new(state)
	player.state.status = waiting
	player.state.current = 0
	player.state.queue = make([]string, 0)
	wd, err := os.Getwd()
	if err == nil && strings.HasSuffix(wd, "music_player") {
		player.playlistsDir = "player/playlists/"
	} else {
		player.playlistsDir = "playlists/"
	}
	return nil
}

// waitEnd is used to wait the end of playing queue
func (player *musicPlayer) waitEnd() {
	player.playQueueMutex.Lock()
	defer player.playQueueMutex.Unlock()
}

// playSingleFile plays single file
// Returns error if file could not be played
func (player *musicPlayer) playSingleFile(filename string, trim float64, ch chan error) error {
	// Open the input file (with default parameters)
	in := sox.OpenRead(filename)
	if in == nil {
		err := errors.New(no_sox_in_msg)
		if ch != nil {
			ch <- err
		}
		return err
	}
	defer in.Release()

	// Open the output device: Specify the output signal characteristics.
	// Since we are using only simple effects, they are the same as the
	// input file characteristics.
	// Using "alsa" or "pulseaudio" should work for most files on Linux.
	// "coreaudio" for OSX
	// On other systems, other devices have to be used.
	out := sox.OpenWrite("default", in.Signal(), nil, "alsa")
	if out == nil {
		out = sox.OpenWrite("default", in.Signal(), nil, "pulseaudio")
		if out == nil {
			out = sox.OpenWrite("default", in.Signal(), nil, "coreaudio")
			if out == nil {
				out = sox.OpenWrite("default", in.Signal(), nil, "waveaudio")
				if out == nil {
					err := errors.New(no_sox_out_msg)
					if ch != nil {
						ch <- err
					}
					return err
				}
			}
		}
	}
	// It's observed that sox crashes at this step sometimes
	defer out.Release()

	if ch != nil {
		ch <- nil
	}

	// Create an effects chain: Some effects need to know about the
	// input or output encoding so we provide that information here.
	chain := sox.CreateEffectsChain(in.Encoding(), out.Encoding())
	defer chain.Release()

	// The first effect in the effect chain must be something that can
	// source samples; in this case, we use the built-in handler that
	// inputs data from an audio file.
	e := sox.CreateEffect(sox.FindEffect("input"))
	e.Options(in)
	// This becomes the first "effect" in the chain
	chain.Add(e, in.Signal(), in.Signal())
	e.Release()

	if trim > 0 {
		interm_signal := in.Signal().Copy()

		e = sox.CreateEffect(sox.FindEffect("trim"))
		e.Options(strconv.FormatFloat(trim, 'f', 2, 64))
		chain.Add(e, interm_signal, in.Signal())
		e.Release()
	}

	// The last effect in the effect chain must be something that only consumes
	// samples; in this case, we use the built-in handler that outputs data.
	e = sox.CreateEffect(sox.FindEffect("output"))
	e.Options(out)
	chain.Add(e, in.Signal(), in.Signal())
	e.Release()

	player.Lock()
	player.state.chain = chain
	player.state.status = playing
	player.state.startTime = time.Now()
	if trim > 0 {
		var milis int64 = int64(-trim * 1000)
		player.state.startTime = player.state.startTime.Add(time.Duration(milis) * time.Millisecond)
	}
	player.Unlock()

	// Flow samples through the effects processing chain until EOF is reached.
	// Flow process is not locked as it must be possible to delete chain effects
	// while Flow is being executed
	// note: sox crashes at this step sometimes(rarely)
	chain.Flow()

	player.Lock()
	if player.state.status == playing {
		player.state.status = waiting
	}
	player.Unlock()

	return nil
}

// play plays a file, directory or playlists
// Returns error if nothing is to be played
func (player *musicPlayer) play(playItem string) ([]string, error) {
	player.stop()
	player.Lock()

	items, err := player.addPlayItem(playItem)
	player.Unlock()

	// play all items
	ch := make(chan error)
	defer close(ch)
	if err == nil {
		go player.playQueue(0, ch)
		err = <-ch
	}
	return items, err
}

// addPlayItem adds a file, directory or playlist to the play queue
// Returns the names of the added songs or error if nothing was added
func (player *musicPlayer) addPlayItem(playItem string) ([]string, error) {
	// is it file or directory
	fileInfo, err := os.Stat(playItem)
	if os.IsNotExist(err) {
		//try it for a playlist
		fileInfo, err = os.Stat(player.playlistsDir + playItem)
		if os.IsNotExist(err) {
			return nil, errors.New(file_not_found_msg)
		}
		playItem = player.playlistsDir + playItem
	}

	items := make([]string, 0)

	switch mode := fileInfo.Mode(); {
	case mode.IsDir():
		d, err := os.Open(playItem)
		if err != nil {
			return nil, errors.New(file_not_found_msg)
		}
		defer d.Close()
		files, err := d.Readdir(-1)
		if err != nil {
			return nil, errors.New(file_not_found_msg)
		}
		prefix := playItem
		if !strings.HasSuffix(playItem, "/") {
			prefix = prefix + "/"
		}
		for _, file := range files {
			if file.Mode().IsRegular() {
				added := player.addRegularFile(prefix + file.Name())
				items = append(items, added...)
			}
		}
	case mode.IsRegular():
		added := player.addRegularFile(playItem)
		items = append(items, added...)

	}
	if len(items) == 0 {
		return nil, errors.New(format_not_supported_msg)
	}
	return items, nil
}

// addRegularFile adds a file or playlist items to the play queue
// Skips the non supported files
// Returns the names of the added files
func (player *musicPlayer) addRegularFile(playItem string) []string {
	items := make([]string, 0)
	if strings.HasSuffix(playItem, playlistsExtension) {
		file, err := os.Open(playItem)
		if err == nil {
			lastSlash := strings.LastIndex(playItem, "/")
			path := ""
			if lastSlash > 0 {
				path = playItem[:lastSlash+1]
			}
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if len(line) > 0 && !strings.HasPrefix(line, "#") {
					if !strings.Contains(line, "/") {
						line = path + line
					}
					name, err := player.addFile(line)
					// if file is not suported - simply skip it
					if err == nil {
						items = append(items, name)
					}
				}
			}
		}
	} else {
		name, err := player.addFile(playItem)
		// if file is not suported - simply skip it
		if err == nil {
			items = append(items, name)
		}
	}
	return items
}

// addFile adds a single file to the player queue
// Checks if file type is supported
func (player *musicPlayer) addFile(fileName string) (string, error) {
	if !isSupportedType(fileName) {
		return "", errors.New(format_not_supported_msg)
	}
	player.state.queue = append(player.state.queue, fileName)
	return fileName, nil
}

// isSupportedType checks if the file type is supported
func isSupportedType(fileName string) bool {
	parts := strings.Split(fileName, ".")
	extension := parts[len(parts)-1]
	supported := false
	if len(parts) > 1 && len(extension) > 0 {
		for _, el := range supportedExtensions {
			if extension == el {
				supported = true
				break
			}
		}
	}
	return supported
}

// playQueue plays the songs in the queue from the current one on
// Trims the song if it was paused
func (player *musicPlayer) playQueue(trim float64, ch chan error) {
	player.playQueueMutex.Lock()
	defer player.playQueueMutex.Unlock()
	play := true
	first := true
	player.Lock()
	player.state.status = waiting
	player.Unlock()
	for play {
		var fileName string
		player.Lock()
		if player.state.status == paused {
			play = false
		} else {
			index := player.state.current
			if index < len(player.state.queue) {
				fileName = player.state.queue[index]
			} else {
				play = false
				player.state.current = 0
				player.state.status = waiting
			}
		}
		player.Unlock()

		if play && len(fileName) > 0 {
			fmt.Println("play queue - song to be played ", fileName)
			if first {
				first = false
				player.playSingleFile(fileName, trim, ch)
			} else {
				player.playSingleFile(fileName, trim, nil)
			}
			trim = 0

			player.Lock()
			if player.state.status == waiting {
				player.state.current += 1
			}
			player.Unlock()
		}
	}
}

// pause pauses the playback
// Returns the name of the paused song or error if player was playing nothing
func (player *musicPlayer) pause() (string, error) {
	player.Lock()
	defer player.Unlock()

	if player.state.status != playing {
		return "", errors.New(cannot_pause_msg)
	}
	player.stopFlow()
	return player.state.queue[player.state.current], nil
}

// stopFlow deletes all effects in the chain so that flow stops
func (player *musicPlayer) stopFlow() {
	// Warning: never call this if the player is not locked
	player.state.chain.DeleteAll()
	player.state.durationPaused = time.Since(player.state.startTime)
	player.state.status = paused
}

// resume resumes the playback
// Returns  the name of the resumed song or error is player was not paused
func (player *musicPlayer) resume() (string, error) {
	player.Lock()

	if player.state.status != paused || player.state.current >= len(player.state.queue) {
		player.Unlock()
		return "", errors.New(cannot_resume_msg)
	}

	songToResume := player.state.queue[player.state.current]
	pausedDuration := player.state.durationPaused
	player.Unlock()

	ch := make(chan error)
	defer close(ch)
	go player.playQueue(pausedDuration.Seconds(), ch)
	err := <-ch
	return songToResume, err
}

// addToQueue adds a song to the queue
// Starts playing if player is in waiting state
// Returns added songs or error if nothing was added
func (player *musicPlayer) addToQueue(playItem string) ([]string, error) {
	player.Lock()

	items, err := player.addPlayItem(playItem)

	//start playing if in Waiting status
	if player.state.status == waiting {
		player.Unlock()
		ch := make(chan error)
		defer close(ch)
		if err == nil {
			go player.playQueue(0, ch)
			err = <-ch
		}
	} else {
		player.Unlock()
	}

	return items, err
}

// stop stops the playback and clears the player's state
func (player *musicPlayer) stop() {
	player.Lock()
	defer player.Unlock()
	if player.state != nil {
		if player.state.chain != nil {
			player.state.chain.DeleteAll()
		}
		player.state.status = paused
		player.state.current = 0
		player.state.queue = make([]string, 0)
	}
}

// next plays the next song from the queue
// Returns the name of the song or error if there is no next song
func (player *musicPlayer) next() (string, error) {
	player.Lock()
	var songToResume string
	if player.state.current < len(player.state.queue)-1 {
		if player.state.status == playing {
			player.stopFlow()
		}
		player.state.current += 1
		songToResume = player.state.queue[player.state.current]

	} else {
		player.Unlock()
		return songToResume, errors.New(cannot_next_msg)
	}

	player.Unlock()
	ch := make(chan error)
	defer close(ch)
	go player.playQueue(0, ch)
	err := <-ch

	return songToResume, err
}

// previous plays the previous song from the queue
// Returns the name of the song or an error if there is no previous song
func (player *musicPlayer) previous() (string, error) {
	player.Lock()
	var songToResume string
	if player.state.current > 0 {
		if player.state.status == playing {
			player.stopFlow()
		}
		player.state.current -= 1
		songToResume = player.state.queue[player.state.current]
	} else {
		player.Unlock()
		return songToResume, errors.New(cannot_previous_msg)
	}

	player.Unlock()
	ch := make(chan error)
	defer close(ch)
	go player.playQueue(0, ch)
	err := <-ch

	return songToResume, err
}

// getCurrentSongInfo gets the name of the current song
// Returns the name of the current song or error if there is no current song
func (player *musicPlayer) getCurrentSongInfo() (string, error) {
	player.Lock()
	defer player.Unlock()
	if player.state.current < len(player.state.queue) {
		return player.state.queue[player.state.current], nil
	}
	return "", errors.New(cannot_get_info_msg)
}

// saveAsPlaylist saves the contents of the queue as a playlist
// Returns the name of the playlist or an error if the playlist could not be saved
func (player *musicPlayer) saveAsPlaylist(playlistName string) (string, error) {
	songs, err := player.getQueueInfo()
	if err != nil {
		return "", errors.New(cannot_save_empty_queue_msg)
	}

	if strings.Contains(playlistName, "/") {
		return "", errors.New(cannot_save_playlist_msg)
	}

	// check the directory
	_, err = os.Stat(player.playlistsDir)
	if os.IsNotExist(err) {
		os.Mkdir(player.playlistsDir, 0777)
	}

	// now create the file
	name := playlistName
	if !strings.HasSuffix(playlistName, playlistsExtension) {
		name = playlistName + playlistsExtension
	}
	file, err := os.Create(player.playlistsDir + name)
	if err != nil {
		return "", errors.New(cannot_save_playlist_msg)
	}
	// https://en.wikipedia.org/wiki/M3U#File_format
	// using the non extended format
	for _, song := range songs {
		file.WriteString(song)
		file.WriteString("\n")
	}
	return name, nil
}

// listPlaylists returns all playlist names from the dedicated directory
// or error if no playlists are found
func (player *musicPlayer) listPlaylists() ([]string, error) {
	player.Lock()
	defer player.Unlock()
	//only the playlists in playlist directory is exposed
	fileInfo, err := os.Stat(player.playlistsDir)
	if os.IsNotExist(err) || !fileInfo.IsDir() {
		return nil, errors.New(playlist_not_found_msg)
	}
	playlists := make([]string, 0)
	d, err := os.Open(player.playlistsDir)
	if err != nil {
		return nil, errors.New(playlist_not_found_msg)
	}
	defer d.Close()
	files, err := d.Readdir(-1)
	if err != nil {
		return nil, errors.New(playlist_not_found_msg)
	}

	for _, file := range files {
		if file.Mode().IsRegular() && strings.HasSuffix(file.Name(), playlistsExtension) {
			playlists = append(playlists, file.Name())
		}
	}
	if len(playlists) == 0 {
		return nil, errors.New(playlist_not_found_msg)
	}
	return playlists, nil
}

// getQueueInfo gets the queue info
// Returns all filenames that are currently in the queue or error if queue is empty
func (player *musicPlayer) getQueueInfo() ([]string, error) {
	player.Lock()
	defer player.Unlock()
	if len(player.state.queue) == 0 {
		return nil, errors.New(cannot_get_queue_info_msg)
	}
	//make a copy to the queue
	copy := make([]string, 0, len(player.state.queue))
	for _, el := range player.state.queue {
		copy = append(copy, el)
	}
	return copy, nil
}
