# music_player

    Note: music_player is still under development and is NOT working

## What is music_player?

music_player is a music player and a RESTful web service that provides control for the playback.

## How do I run music_player?

music_player uses SoX internally (http://sox.sourceforge.net/)
So you should install it first.

*Linux:*

  Installation depend on your distribution. You may try:
    
  **yum install sox**
      
  or:
    
  **apt-get install sox**

*OSX:*

  The easiest way is to use Homebrew (http://brew.sh/)

  **brew install sox**
   
To get this project execute:

  **go get github.com/katya-spasova/music_player**

Start the service:
  **go run start_service.go**
    
## What are the supported music formats?

music_player supports mp3 so far. Others may be added in the future.
    
## How do I use music_player?

music_player comes with a client, called playback_control. Go to music_player directory and 
start it. Check the --help to see how it is used.

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

### JSON Response
The json response in case the operation is successful look similar to the following example:

~~~json
{
   "Code": 0,
   "Message": "Success",
   "Info": "Queue content",
   "Data": [
      "test_sounds/beep9.mp3",
      "test_sounds/beep28.mp3",
      "test_sounds/beep36.mp3"
   ]
}
~~~

The json response in case the operation fails looks similar to:

~~~json
{
   "Code": 11,
   "Message": "Cannot play previous song. No previous song in queue"
}
~~~
    
### Codes used in the json response

| Code | Message |
| --- | --- |
| 0 | Success |
| 1 | Failed to initialize SoX |
| 2 | SoX failed to open input file |
| 3 | Sox failed to open output device |
| 4 | File cannot be found |
| 5 | Playlist cannot be found |
| 6 | Currently there are no saved playlists |
| 7 | Format is not supported |
| 8 | Cannot pause. No song is playing |
| 9 | Cannot resume. No song was paused |
| 10 | Cannot play next song. No next song in queue |
| 11 | Cannot play previous song. No previous song in queue |
| 12 | There is no current song in the queue |
| 13 | Cannot save playlist |
| 14 | Queue is empty and cannot be saved as playlist |
| 15 | Cannot get queue info. Queue is empty |

## Why would I use music_player?

Ever happended to you to listen to music and the computer you're working/playing on is not the 
one that plays the music? Your mobile phone is very likely to be a computer too.
You had to switch the computer to change the song you're listening to.
With music_player you can do it from any device as long as your computers are connected 
to the internet (or are in the same local network).
    
## External sources?

music_player uses **github.com/krig/go-sox** and **goji.io**