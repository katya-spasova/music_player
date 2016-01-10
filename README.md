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

  ** brew install sox**
   
To get this project execute:

  ** go get github.com/katya-spasova/music_player **
    
And run it from music_player/player directory execute:

  ** go run service.go player.go
    
## What are the supported music formats?

music_player supports what SoX supports. run **sox -h** and check **AUDIO FILE FORMATS** section
    
## How do I use music_player?

music_player comes with a client, called playback_control. Go to music_player directory and 
start it. Check the --help to see how it is used.

You can use music_player by directly sending HTTP request to it. Check below to see the API.

And of course you can write your own client for any platform you like.
    
## What's the web service API?

**GET host:8765/** - checks if the service is alive

**PUT host:8765/play/<filename/directory/playlist>** - plays music from file, directory or playlist

**POST host:8765/pause** - pauses the playback

**POST host:8765/resume** - resumes the playback

**PUT host:8765/stop** - stops the playback (cannot be resumed)

**POST host:8765/next** - plays the next song

**POST host:8765/previous** - plays the previous song

**PUT host:8765/songinfo** - returns info about the current song

**POST host:8765/add/<filename/directory/playlist>** - add music to the play queue from file, directory, playlist

**PUT host:8765/save/<playlist>** - saves the play queue to a playlist

**GET host :8765/playlists** - returns a list of all saved playlists

    //todo: describe the response json and the error codes
    

## Why would I use music_player?

Ever happended to you to listen to music and the computer you're working/playing on is not the 
one that plays the music? Your mobile phone is very likely to be a computer too.
You had to switch the computer to change the song you're listening to.
With music_player you can do it from any device as long as your computers are connected 
to the internet (or are in the same local network).
    
## External sources?

music_player uses **github.com/krig/go-sox** and **goji.io**