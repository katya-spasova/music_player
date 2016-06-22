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
    queueInfo();
}

function addSong() {
    sendToPlayer("POST", "add/");
    currentSong();
    queueInfo();
}

function stopSong() {
    sendToPlayer("PUT", "stop");
    currentSong();
    queueInfo();
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
    currentSong();
    queueInfo();
}

function resumeSong() {
    sendToPlayer("POST", "resume");
    currentSong();
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

function jumpToSong(songIndex) {
    sendToPlayer("POST", "jump/" + songIndex);
    currentSong();
}

function sendToPlayer(method, action, afterFunction, cb) {
    var nameElement = document.getElementById("name");
    var name = isNameApplicable(action) ? encodeURIComponent(nameElement.value) : "";
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
        afterFunction && afterFunction(xhttp, getElementId(action), cb);
    };
    xhttp.open(method, action.concat(name), true);
    xhttp.send();
}

function isNameApplicable(action) {
    return action == "play/" || action == "add/" || action == "save/";
}

function updateContent(xhttp, elementId, cb) {
    if (xhttp.readyState == 4 && xhttp.status == 200) {
        updateContentImpl(elementId, xhttp.responseText);
        cb && cb();
    }
}

function updateContentImpl(elementId, responseText) {
    // parse and produce html
    var res = JSON.parse(responseText);
    var content = "";
    if (res["Code"] > 0) {
        content = res["Message"];
    } else if (elementId == "queue") {
        var songs = res["Data"];
        if (typeof songs != "undefined") {
            for (var i = 0; i < songs.length; i++) {
                content = content + "<div><a href='#' id='" + i + "'>" + songs[i] + "</a></div>"
            }
        }
    } else {
        content = res["Data"];
    }

    // update response
    document.getElementById(elementId).innerHTML = content;

    // add event listeners
    if (res["Code"] == 0 && elementId == "queue") {
        songs = res["Data"];
        if (typeof songs != "undefined") {
            for (var j = 0; j < songs.length; j++) {
                document.getElementById(j.toString()).addEventListener("click", function (event) {
                    event.preventDefault();
                    jumpToSong(event.target.id);
                });
            }
        }
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
