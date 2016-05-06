function init() {
    document.getElementById("play").addEventListener("click", function(event){
        event.preventDefault();
        playSong();
    });
    document.getElementById("add").addEventListener("click", addSong);
    document.getElementById("stop").addEventListener("click", stopSong);
    document.getElementById("previous").addEventListener("click", previousSong);
    document.getElementById("next").addEventListener("click", nextSong);
    document.getElementById("pause").addEventListener("click", pauseSong);
    document.getElementById("playlist").addEventListener("click", getPlaylist);
}


function playSong() {
    var nameElement = document.getElementById("name");
    var name = encodeURIComponent(nameElement.value);
    var xhttp = new XMLHttpRequest();
    xhttp.open("PUT", "play/".concat(name), true);
    xhttp.send();
}

document.addEventListener('DOMContentLoaded', function() {
   init();
}, false);
