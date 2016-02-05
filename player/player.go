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

const playlistsDir = "playlists" + os.PathSeparator
const playlistsExtension = ".m3u"

type Player struct {
	WG *sync.WaitGroup
	sync.Mutex
	State *State
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
	player.State = new(State)
	player.State.status = Waiting
	player.State.current = 0
	player.State.queue = make(string, 0)
	return nil
}

// Clears resources
func (player *Player) clear() {
	player.Lock()
	defer player.Unlock()
	sox.Quit()
	if player.State.in != nil {
		player.State.in.Release()
	}

	if player.State.out != nil {
		player.State.out.Release()
	}

	if player.State.chain != nil {
		player.State.chain.Release()
	}
}

// Plays single file
func (player *Player) playSingleFile(filename string, trim float32) error {
	// Open the input file (with default parameters)
	in := sox.OpenRead(filename)
	if in == nil {
		if player.WG != nil {
			player.WG.Done()
		}
		return errors.New(no_sox_in_msg)
	}
	player.State.in = in

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
					if player.WG != nil {
						player.WG.Done()
					}
					return errors.New(no_sox_out_msg)
				}
			}
		}
	}
	player.State.out = out

	// Create an effects chain: Some effects need to know about the
	// input or output encoding so we provide that information here.
	chain := sox.CreateEffectsChain(in.Encoding(), out.Encoding())
	player.State.chain = chain

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

	player.State.status = Playing
	player.State.startTime = time.Now()

	// Flow samples through the effects processing chain until EOF is reached.
	go func(wg *sync.WaitGroup) {
		chain.Flow()
		if wg != nil {
			wg.Done()
		}
		player.State.status = Waiting
	}(player.WG)

	return nil
}

func (player *Player) play(playItem string) ([]string, error) {
	player.stop()
	player.Lock()

	items, err := player.addPlayItem(playItem)
	player.Unlock()

	// play all items
	if err != nil {
		go player.playQueue(0)
	}
	return items, err
}

func (player *Player) addPlayItem(playItem string) ([]string, error) {
	// is it file or directory
	fileInfo, err := os.Stat(playItem)
	if os.IsNotExist(err) != nil {
		//try it for a playlist
		fileInfo, err = os.Stat(playlistsDir + playItem + ".m3u")
		if os.IsNotExist(err) {
			return nil, errors.New(playItem + " does not exist")
		}
	}

	items := make([]string, 0)

	switch mode := fileInfo.Mode(); {
	case mode.IsDir():

		d, err := os.Open(playItem)
		if err != nil {
			return nil, err
		}
		defer d.Close()
		files, err := d.Readdir(-1)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			if file.Mode().IsRegular() {
				added := player.addRegularFile(file.Name())
				items = append(items, added...)
			}
		}
	case mode.IsRegular():
		added := player.addRegularFile(playItem)
		items = append(items, added...)

	}
	return items, nil
}

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
func (player *Player) addFile(file string) (string, error) {
	//todo: check if file type is supported
	player.State.queue = append(player.State.queue, file)
	return file, nil
}

func (player *Player) playQueue(trim float32) {
	play := true
	for play {
		var fileName string
		player.Lock()
		index := player.State.current
		if index < len(player.State.queue) {
			fileName = player.State.queue[index]
		} else {
			play = false
			player.State.current = 0
			player.State.status = Waiting
		}
		player.Unlock()

		if play && len(fileName) > 0 {
			player.playSingleFile(fileName, trim)
		}
		player.Lock()
		player.State.current += 1
		player.Unlock()
	}
}

func (player *Player) pause() error {
	player.Lock()
	defer player.Unlock()

	if player.State.status != Playing {
		return errors.New("Cannot pause. Player is not playing")
	}

	player.State.chain.DeleteAll()
	player.State.durationPaused = time.Since(player.State.startTime)
	player.State.status = Paused

	return nil
}

func (player *Player) resume() (string, error) {
	player.Lock()

	if player.State.status != Paused {
		return "", errors.New("Cannot resume. Player is not paused")
	}

	songToResume := player.State.queue[player.State.current]
	pausedDuration := player.State.durationPaused
	player.Unlock()

	go player.playQueue(pausedDuration.Seconds())
	return songToResume, nil
}

func (player *Player) addToQueue(playItem string) ([]string, error) {
	player.Lock()

	items, err := player.addPlayItem(playItem)

	//start playing if in Waiting status
	if player.State.status == Waiting {
		player.Unlock()
		if err != nil {
			go player.playQueue(0)
		}
	} else {
		player.Unlock()
	}

	return items, err
}

func (player *Player) stop() {
	player.Lock()
	defer player.Unlock()
	player.State.chain.DeleteAll()
	player.State.status = Waiting
	player.State.current = 0
	player.State.queue = make(string, 0)
}

func (player *Player) next() (string, error) {
	player.Lock()
	var songToResume string
	if player.State.current < len(player.State.queue)-1 {
		player.State.current += 1
		player.State.chain.DeleteAll()
		songToResume = player.State.queue[player.State.current]

	} else {
		player.Unlock()
		return songToResume, errors.New("No next song in queue")
	}

	player.Unlock()
	go player.playQueue(0)

	return songToResume, nil
}

func (player *Player) previous() error {
	player.Lock()
	var songToResume string
	if player.State.current > 0 {
		player.State.current -= 1
		player.State.chain.DeleteAll()
		songToResume = player.State.queue[player.State.current]
	} else {
		player.Unlock()
		return songToResume, errors.New("No previous song in queue")
	}

	player.Unlock()
	go player.playQueue(0)

	return songToResume, nil
}

func (player *Player) getCurrentSongInfo() (string, error) {
	player.Lock()
	defer player.Unlock()
	if player.State.current < len(player.State.queue) {
		return player.State.queue[player.State.current], nil
	}
	return "", errors.New("No current song found")
}

func (player *Player) saveAsPlaylist(playlistName string) (string, error) {
	songs, err := player.getQueueInfo()
	if err != nil {
		return "", err
	}
	// check the directory
	fileInfo, err := os.Stat(playlistsDir)
	if os.IsNotExist(err) {
		os.Mkdir(playlistsDir, 0777)
	} else if !fileInfo.IsDir() {
		return "", errors.New("Existing file in place of directory")
	}

	// now create the file
	name := playlistName
	if !strings.HasSuffix(playlistName, playlistsExtension) {
		name = playlistName + playlistsExtension
	}
	file, err := os.Create(name)
	if err != nil {
		return "", err
	}
	// https://en.wikipedia.org/wiki/M3U#File_format
	// using the non extended format
	for song := range songs {
		file.WriteString(song)
		file.WriteString("\n")
	}
	return name, nil
}

func (player *Player) listPlaylists() ([]string, error) {
	player.Lock()
	defer player.Unlock()
	//only the playlists in playlist directory is exposed
	fileInfo, err := os.Stat(playlistsDir)
	if os.IsNotExist(err) || !fileInfo.IsDir() {
		return nil, errors.New("no playlists dir")
	}
	playlists := make([]string, 0)
	d, err := os.Open(playlistsDir)
	if err != nil {
		return nil, err
	}
	defer d.Close()
	files, err := d.Readdir(-1)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.Mode().IsRegular() && strings.HasSuffix(file.Name(), playlistsExtension) {
			playlists = append(playlists, file.Name())
		}
	}
	return "", nil
}

func (player *Player) getQueueInfo() ([]string, error) {
	player.Lock()
	defer player.Unlock()
	if len(player.State.queue) == 0 {
		return nil, errors.New("Empty queue")
	}
	//make a copy to the queue
	copy := make([]string, 0, len(player.State.queue))
	for el := range player.State.queue {
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
