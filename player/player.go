package player

import "github.com/krig/go-sox"
import (
	"bufio"
	"errors"
	"os"
	"strings"
	"sync"
	"time"
)

const playlistsDir string = "playlists/"
const playlistsExtension = ".m3u"

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

type Player struct {
	wg *sync.WaitGroup
	sync.Mutex
	state *State
}

// InOut struct holds the state of the player
type State struct {
	in             *sox.Format
	out            *sox.Format
	chain          *sox.EffectsChain
	status         int
	startTime      time.Time
	durationPaused time.Duration
	queue          []string
	current        int
}

const (
	Playing = iota
	Paused
	Waiting
)

// Starts sox and the initialises player's state
func (player *Player) init() error {
	player.Lock()
	defer player.Unlock()
	if !sox.Init() {
		return errors.New(no_sox_msg)
	}
	player.state = new(State)
	player.state.status = Waiting
	player.state.current = 0
	player.state.queue = make([]string, 0)
	return nil
}

// Clears resources
func (player *Player) clear() {
	player.Lock()
	defer player.Unlock()
	sox.Quit()
	if player.state.in != nil {
		player.state.in.Release()
	}

	if player.state.out != nil {
		player.state.out.Release()
	}

	if player.state.chain != nil {
		player.state.chain.Release()
	}
}

// Plays single file
// Returns error if file could not be played
func (player *Player) playSingleFile(filename string, trim float64) error {
	// Open the input file (with default parameters)
	in := sox.OpenRead(filename)
	if in == nil {
		//		if player.wg != nil {
		//			player.wg.Done()
		//		}
		return errors.New(no_sox_in_msg)
	}

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
					if player.wg != nil {
						player.wg.Done()
					}
					return errors.New(no_sox_out_msg)
				}
			}
		}
	}

	// Create an effects chain: Some effects need to know about the
	// input or output encoding so we provide that information here.
	chain := sox.CreateEffectsChain(in.Encoding(), out.Encoding())

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
		e.Options(trim) //todo: try with float ?!?
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

	player.state.in = in
	player.state.out = out
	player.state.chain = chain
	player.state.status = Playing
	player.state.startTime = time.Now()
	player.Unlock()

	// Flow samples through the effects processing chain until EOF is reached.
	//	go func(wg *sync.WaitGroup) {
	chain.Flow()
	player.Lock()
	player.state.status = Waiting
	player.Unlock()
	//		if wg != nil {
	//			wg.Done()
	//		}
	//	}(player.wg)

	return nil
}

// Plays a file, directory or playlists
// Returns error if nothing is to be played
func (player *Player) play(playItem string) ([]string, error) {
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

// Adds a file, directory or playlist to the play queue
// Returns the names of the added songs or error if nothing was added
func (player *Player) addPlayItem(playItem string) ([]string, error) {
	// is it file or directory
	fileInfo, err := os.Stat(playItem)
	if os.IsNotExist(err) {
		//try it for a playlist
		fileInfo, err = os.Stat(playlistsDir + playItem + playlistsExtension)
		if os.IsNotExist(err) {
			return nil, errors.New(file_not_found_msg)
		}
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

// Adds a file or playlist items to the play queue
// Skips the non supported files
// Returns the names of the added files
func (player *Player) addRegularFile(playItem string) []string {
	items := make([]string, 0)
	if strings.HasSuffix(playItem, playlistsExtension) {
		file, err := os.Open(playItem)
		if err == nil {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if len(line) > 0 && !strings.HasPrefix(line, "#") {
					name, err := player.addFile(playItem)
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

// Adds a single file to the player queue
// Checks if file type is supported
func (player *Player) addFile(fileName string) (string, error) {
	if !isSupportedType(fileName) {
		return "", errors.New(format_not_supported_msg)
	}
	player.state.queue = append(player.state.queue, fileName)
	return fileName, nil
}

// Checks if the file type is supported
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

// Plays the songs in the queue from the current one on
// Trims the song if it was paused
func (player *Player) playQueue(trim float64, ch chan error) {
	play := true
	first := true
	for play {
		var fileName string
		player.Lock()
		index := player.state.current
		if index < len(player.state.queue) {
			fileName = player.state.queue[index]
		} else {
			play = false
			player.state.current = 0
			player.state.status = Waiting
		}
		player.Unlock()

		if play && len(fileName) > 0 {
			err := player.playSingleFile(fileName, trim)
			if first {
				first = false
				ch <- err
			}
		}
		player.Lock()
		player.state.current += 1
		player.Unlock()
	}
}

// Pauses the playback
// Returns the name of the paused song or error if player was playing nothing
func (player *Player) pause() (string, error) {
	player.Lock()
	defer player.Unlock()

	if player.state.status != Playing {
		return "", errors.New(cannot_pause_msg)
	}

	player.state.chain.DeleteAll()
	player.state.durationPaused = time.Since(player.state.startTime)
	player.state.status = Paused

	return player.state.queue[player.state.current], nil
}

// Resumes the playback
// Returns  the name of the resumed song or error is player was not paused
func (player *Player) resume() (string, error) {
	player.Lock()

	if player.state.status != Paused {
		return "", errors.New(cannot_pause_msg)
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

// Adds a song to the queue
// Starts playing if player is in waiting state
// Returns added songs or error if nothing was added
func (player *Player) addToQueue(playItem string) ([]string, error) {
	player.Lock()

	items, err := player.addPlayItem(playItem)

	//start playing if in Waiting status
	if player.state.status == Waiting {
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

// Stops the playback and clears the player's state
func (player *Player) stop() {
	player.Lock()
	defer player.Unlock()
	if player.state != nil {
		if player.state.chain != nil {
			player.state.chain.DeleteAll()
		}
		player.state.status = Waiting
		player.state.current = 0
		player.state.queue = make([]string, 0)
	}
}

// Plays the next song from the queue
// Returns the name of the song or error if there is no next song
func (player *Player) next() (string, error) {
	player.Lock()
	var songToResume string
	if player.state.current < len(player.state.queue)-1 {
		player.state.current += 1
		player.state.chain.DeleteAll()
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

// Plays the previous song from the queue
// Returns the name of the song or an error if there is no previous song
func (player *Player) previous() (string, error) {
	player.Lock()
	var songToResume string
	if player.state.current > 0 {
		player.state.current -= 1
		player.state.chain.DeleteAll()
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

// Gets the name of the current song
// Returns the name of the current song or error if there is no current song
func (player *Player) getCurrentSongInfo() (string, error) {
	player.Lock()
	defer player.Unlock()
	if player.state.current < len(player.state.queue) {
		return player.state.queue[player.state.current], nil
	}
	return "", errors.New(cannot_get_info_msg)
}

// Saves the contents of the queue as a playlist
// Returns the name of the playlist or an error if the playlist could not be saved
func (player *Player) saveAsPlaylist(playlistName string) (string, error) {
	songs, err := player.getQueueInfo()
	if err != nil {
		return "", errors.New(cannot_save_playlist_msg)
	}
	// check the directory
	fileInfo, err := os.Stat(playlistsDir)
	if os.IsNotExist(err) {
		os.Mkdir(playlistsDir, 0777)
	} else if !fileInfo.IsDir() {
		return "", errors.New(cannot_save_playlist_msg)
	}

	// now create the file
	name := playlistName
	if !strings.HasSuffix(playlistName, playlistsExtension) {
		name = playlistName + playlistsExtension
	}
	file, err := os.Create(name)
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

// Returns all playlist names from the dedicated directory
// or error if no playlists are found
func (player *Player) listPlaylists() ([]string, error) {
	player.Lock()
	defer player.Unlock()
	//only the playlists in playlist directory is exposed
	fileInfo, err := os.Stat(playlistsDir)
	if os.IsNotExist(err) || !fileInfo.IsDir() {
		return nil, errors.New(playlist_not_found_msg)
	}
	playlists := make([]string, 0)
	d, err := os.Open(playlistsDir)
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

// Gets the queue info
// Returns all filenames that are currently in the queue or error if queue is empty
func (player *Player) getQueueInfo() ([]string, error) {
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

//func main() {
//	wg := sync.WaitGroup{}
//	player = Player{WG:&wg}
//	err := player.init()
//	if (err != nil) {
//		return
//	}
//	fmt.Println("Adding")
//	wg.Add(1)
//	player.playSingleFile("/Users/katyaspasova/Music/Vampolka.mp3")
//	wg.Wait()
//	player.clear()
//}
