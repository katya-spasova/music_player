function init() {
    document.getElementById("play").addEventListener("click", function(event){
        event.preventDefault();
        playSong();
    });
    document.getElementById("add").addEventListener("click", function(event){
        event.preventDefault();
        addSong();
    });
    document.getElementById("stop").addEventListener("click", function(event){
        event.preventDefault();
        stopSong();
    });
    document.getElementById("previous").addEventListener("click", function(event){
        event.preventDefault();
        previousSong();
    });
    document.getElementById("next").addEventListener("click", function(event){
        event.preventDefault();
        nextSong();
    });
    document.getElementById("pause").addEventListener("click", function(event){
        event.preventDefault();
        pauseSong();
    });
    document.getElementById("resume").addEventListener("click", function(event){
        event.preventDefault();
        resumeSong();
    });
    document.getElementById("playlist").addEventListener("click", function(event){
        event.preventDefault();
        getPlaylists();
    });
    document.getElementById("queueinfo").addEventListener("click", function(event){
        event.preventDefault();
        queueInfo();
    });
    document.getElementById("save").addEventListener("click", function(event){
        event.preventDefault();
        savePlaylist();
    });
}

function playSong() {
    sendToPlayer("PUT", "play/");
    currentSong();
}

function addSong() {
    sendToPlayer("POST", "add/");
    currentSong();
}

function stopSong() {
    sendToPlayer("PUT", "stop");
}

function previousSong() {
    sendToPlayer("POST", "previous");
    currentSong();
}

function nextSong() {
    sendToPlayer("POST", "next");
    currentSong();
}

function pauseSong() {
    sendToPlayer("POST", "pause");
}

function resumeSong() {
    sendToPlayer("POST", "resume");
}

function currentSongPeriodic() {
    sendToPlayer("GET", "songinfo", updateContent,
        function() {
            setTimeout(currentSongPeriodic, 5000);
        }
    )
}

function currentSong() {
    sendToPlayer("GET", "songinfo", updateContent);
}

function queueInfo() {
    sendToPlayer("GET", "queueinfo", updateContent);
}

function getPlaylists() {
    sendToPlayer("GET", "playlists", updateContent)
}

function savePlaylist() {
    sendToPlayer("PUT", "save/");
    getPlaylists();
}

function sendToPlayer(method, action, afterFunction, cb) {
    var nameElement = document.getElementById("name");
    var name = isNameApplicable(action) ? encodeURIComponent(nameElement.value) : "";
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
        afterFunction && afterFunction(xhttp, getElementId(action), cb);
    }
    xhttp.open(method, action.concat(name), true);
    xhttp.send();
}

function isNameApplicable(action) {
    if (action == "play/" || action == "add/" || action == "save/") {
        return true;
    }
    return false;
}

function updateContent(xhttp, elementId, cb) {
    if (xhttp.readyState == 4 && xhttp.status == 200) {
        document.getElementById(elementId).innerHTML = xhttp.responseText;
        cb && cb();
    }
}

function getElementId(action) {
    if (action == "queueinfo") {
        return "queue";
    }

    if (action == "playlists") {
        return "playlists";
    }

    if (action == "songinfo") {
        return "currentSong";
    }
}

document.addEventListener('DOMContentLoaded', function() {
   init();
   currentSongPeriodic();
   getPlaylists();
   queueInfo();
}, false);