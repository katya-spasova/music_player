# music_player

## What is music_player?

music_player is a music player and a RESTful web service that provides control for the playback.

## How do I run music_player?

* **1. music_player uses SoX internally (http://sox.sourceforge.net/). So you should install it first.**

*Linux:*

  Installation depends on your distribution.

  For example on Ubuntu:

~~~sh
  sudo apt-get install sox libsox-fmt-all libsox-dev
~~~

*OSX:*

  The easiest way is to use Homebrew (http://brew.sh/)
~~~sh
  brew install sox --with-libvorbis --with-flac --with-lame
~~~

* **2. To get this project execute:**

~~~sh
   go get github.com/katya-spasova/music_player
~~~

* **3. Start the service:**
~~~sh
   go run start_service.go
~~~

* **4. To run the unit tests**
~~~sh
  cd $GOPATH/src/github.com/katya-spasova/music_player/player/
  go test
~~~

## What are the supported music formats?

music_player will let you work with:
*8svx aif aifc aiff aiffc al amb au avr cdda cdr cvs cvsd cvu dat dvms f32 f4 f64 f8 flac fssd gsm gsrt hcom htk ima ircam la lpc lpc10 lu maud mp2 mp3 nist ogg prc raw s1 s16 s2 s24 s3 s32 s4 s8 sb sf sl sln smp snd sndr sndt sou sox sph sw txw u1 u16 u2 u24 u3 u32 u4 u8 ub ul uw vms voc vox wav wavpcm wve xa*

## How do I use music_player?

music_player comes with a client, called playback_control. Go to music_player/playback_control directory and execute
*go run start_client.go -help* (or build it first if you wish).

You can use music_player by directly sending HTTP request to it. Check below to see the API.

And of course you can write your own client for any platform you like.

## What's the web service API?

| PATH | Description|
| --- | --- |
| GET host:8765/ | checks if the service is alive|
| PUT host:8765/play/<filename/directory/playlist> | plays music from file, directory or playlist |
| POST host:8765/pause | pauses the playback |
| POST host:8765/resume | resumes the playback |
| PUT host:8765/stop | stops the playback (cannot be resumed) |
| POST host:8765/next | plays the next song |
| POST host:8765/previous | plays the previous song |
| GET host:8765/songinfo | returns info about the current song |
| POST host:8765/add/<filename/directory/playlist> | add music to the play queue from file, directory, playlist |
| PUT host:8765/save/<playlist> | saves the play queue to a playlist |
| GET host:8765/playlists | returns a list of all saved playlists |
| GET host:8765/queueinfo | returns list of all songs in the queue |
| POST host:8765/jump/<index> | plays a song with specific index from the queue |

### JSON Response
The json response in case the operation is successful look similar to the following example:

~~~json
{
   "Code": 0,
   "Message": "Queue content",
   "Data": [
      "beep9.mp3",
      "beep28.mp3",
      "beep36.mp3"
   ]
}
~~~

The json response in case the operation fails looks similar to:

~~~json
{
   "Code": 1,
   "Message": "Cannot play previous song. No previous song in queue"
}
~~~

### Codes used in the json response

Code "0" is used for success and "1" failure

| Code | Message |
| --- | --- |
| 0 | Started playing |
| 0 | Added to queue |
| 0 | Song is paused |
| 0 | Song is resumed |
| 0 | Playback is stopped and cleaned |
| 0 | The filename of the current song |
| 0 | The filenames in the current queue |
| 0 | The queue is saved as a playlist |
| 0 | A list of all saved playlists |
| 0 | Queue content |
| 1 | SoX failed to open input file |
| 1 | Sox failed to open output device |
| 1 | File cannot be found |
| 1 | Playlist cannot be found |
| 1 | Currently there are no saved playlists |
| 1 | Format is not supported |
| 1 | Cannot pause. No song is playing |
| 1 | Cannot resume. No song was paused |
| 1 | Cannot play next song. No next song in queue |
| 1 | Cannot play previous song. No previous song in queue |
| 1 | There is no current song in the queue |
| 1 | Cannot save playlist |
| 1 | Queue is empty and cannot be saved as playlist |
| 1 | Cannot get queue info. Queue is empty |
| 1 | Song not available |

## Why would I use music_player?

Ever happended to you to listen to music and the computer you're working/playing on is not the
one that plays the music? Your mobile phone is very likely to be a computer too.
You had to switch the computer to change the song you're listening to.
With music_player you can do it from any device as long as your computers are connected
to the internet (or are in the same local network).

## External sources?

music_player uses **github.com/krig/go-sox** and **goji.io**
