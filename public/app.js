var ws;
var ingame = false;

function createGame() {
    var gameName = document.getElementById("gameName").value;
    callServerCreateGame(gameName); //this func will join game after creating
}

function callServerCreateGame(gameName) {
    var http = new XMLHttpRequest();
    http.open("POST", "/create", true);
    http.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
    http.onreadystatechange = function () {
        if(http.readyState == 4 && http.status == 200) {
            if(http.responseText == "Game already exists") {
                alert(http.responseText);
            } else {
                joinGame();
            }
        }
    }
    http.send("gameName=" + gameName.toUpperCase());
}

function joinGame()
{
    ws = new WebSocket("ws://" + location.host + "/websocket");

    ws.onopen = function() {
        var gameName = document.getElementById("gameName").value;
        var name = document.getElementById("name").value;
        var messageJson = '{"gameName":"' + gameName.toUpperCase() + '", "name":"' + name.toUpperCase() + '"}';
        ws.send(messageJson);
    }

    ws.onmessage = function (evt)
    {
        if(evt.data == "Not enough players" || 
        evt.data == "Game full" ||
        evt.data == "Game started" ||
        evt.data == "Duplicate name" ||
        evt.data == "Game does not exist") {
            alert(evt.data);
        } else if (evt.data == "Joined"){
            ingame = true;
            document.getElementById("gameName").remove();
            document.getElementById("name").remove();
            document.getElementById("createBtn").remove();
            document.getElementById("joinBtn").remove();
        } else if (evt.data == "First") {
            var button = document.createElement("button");
            button.innerHTML = "START";
            button.className = "button";
            button.id = "startBtn";
            button.addEventListener ("click", function() {
                ws.send("start");
            });
            document.getElementsByTagName("body")[0].appendChild(button);
        } else if (evt.data == "Started") {
            document.getElementById("startBtn").remove();
        } else {
            var personList = document.getElementById("personList");
            personList.innerHTML = (evt.data);
        }
    };

    ws.onclose = function()
    {
        if(ingame == true) {
            alert("You have been disconnected");
            location.reload();
        }
    };
}

function sendInvestigationMessage(index) {
    if(confirm("Are you sure you want to investigate?")) {
        ws.send(index);
    }
}