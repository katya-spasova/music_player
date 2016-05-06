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
}

function playSong() {
    sendToPlayer("PUT", "play/", doNothing);
}

function addSong() {
    sendToPlayer("POST", "add/", doNothing);
}

function stopSong() {
    sendToPlayer("PUT", "stop", doNothing);
}

function previousSong() {
    sendToPlayer("POST", "previous", doNothing);
}

function nextSong() {
    sendToPlayer("POST", "next", doNothing);
}

function pauseSong() {
    sendToPlayer("POST", "pause", doNothing);
}

function resumeSong() {
    sendToPlayer("POST", "resume", doNothing);
}

function currentSong() {
    sendToPlayer("GET", "songinfo", doNothing)
}

function queueInfo() {
    sendToPlayer("GET", "queueinfo", updateContent)
}

function getPlaylists() {
    sendToPlayer("GET", "playlists", updateContent)
}

function sendToPlayer(method, action, afterFunction) {
    var nameElement = document.getElementById("name");
    var name = isNameApplicable(action) ? encodeURIComponent(nameElement.value) : "";
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
        afterFunction(xhttp, getElementId(action));
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

function updateContent(xhttp, elementId) {
    if (xhttp.readyState == 4 && xhttp.status == 200) {
        document.getElementById(elementId).innerHTML = xhttp.responseText;
    }
}

function getElementId(action) {
    if (action == "queueinfo") {
        return "queue";
    }

    if (action == "playlists") {
        return "playlists";
    }
}

function doNothing() {
}

document.addEventListener('DOMContentLoaded', function() {
   init();
}, false);
