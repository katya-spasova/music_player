package player

import "github.com/krig/go-sox"
import (
	"errors"
	"sync"
	"time"
)

type Player struct {
	WG    *sync.WaitGroup
	sync.Mutex
	State *State
}

// InOut struct holds the state of the player
type State struct {
	in        *sox.Format
	out       *sox.Format
	chain     *sox.EffectsChain
	status    int
	startTime time.Time
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
func (player *Player) playSingleFile(filename string) error {
	player.Lock()
	defer player.Unlock()
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

	//	interm_signal := in.Signal().Copy()

	// The first effect in the effect chain must be something that can
	// source samples; in this case, we use the built-in handler that
	// inputs data from an audio file.
	e := sox.CreateEffect(sox.FindEffect("input"))
	e.Options(in)
	// This becomes the first "effect" in the chain
	chain.Add(e, in.Signal(), in.Signal())
	e.Release()

	//	e = sox.CreateEffect(sox.FindEffect("trim"))
	//	e.Options("10")
	//	chain.Add(e, interm_signal, in.Signal())
	//	e.Release()

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
