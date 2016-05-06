function init() {
    document.getElementById("play").addEventListener("click", playSong);
    document.getElementById("add").addEventListener("click", addSong);
    document.getElementById("stop").addEventListener("click", stopSong);
    document.getElementById("previous").addEventListener("click", previousSong);
    document.getElementById("next").addEventListener("click", nextSong);
    document.getElementById("pause").addEventListener("click", pauseSong);
    document.getElementById("playlist").addEventListener("click", getPlaylist);
}


function playSong() {
    var name = encodeURIComponent(document.getElementById("name"));
    xhttp.open("POST", "play/".concat(name), true);
    xhttp.send();
}

document.onload = init();

