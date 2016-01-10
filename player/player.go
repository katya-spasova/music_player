package main

import "github.com/krig/go-sox"
import "log"

type Player struct {
	//todo: add queue, timer, io state etc.
}

// InOut struct holds the state of the player
type InOut struct {
	In    *sox.Format
	Out   *sox.Format
	Chain *sox.EffectsChain
}

// Starts sox and the initialises player's state
func (player *Player) init() {
	if !sox.Init() {
		log.Fatal("Failed to initialize SoX")
	}
	//todo: init the state
}

// Stops sox
func (player *Player) clear() {
	sox.Quit()
}

// Plays single file
func (player *Player) playSingleFile(filename string) {
	// Open the input file (with default parameters)
	in := sox.OpenRead(filename)
	if in == nil {
		log.Fatal("Failed to open input file")
	}
	// Close the file before exiting
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
				log.Fatal("Failed to open output device")
			}
		}
	}
	// Close the output device before exiting
	defer out.Release()

	// Create an effects chain: Some effects need to know about the
	// input or output encoding so we provide that information here.
	chain := sox.CreateEffectsChain(in.Encoding(), out.Encoding())
	// Make sure to clean up!
	defer chain.Release()

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

	// Flow samples through the effects processing chain until EOF is reached.
	go chain.Flow()

	// todo: save InOut{In: in, Out:out, Chain: chain}
}
