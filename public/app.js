var ws;
var ingame = false;
var name = "";
var gameName = "";
var isconnected = false;
var rememberInv = -1;

function toggleCreateMode() {
    var createMode = document.getElementById("createSwitch").checked;
    console.log(createMode);
    var joinBtn = document.getElementById("joinBtn");
    if(createMode == 1) {
        joinBtn.value = "CREATE";
        joinBtn.onclick = createGame;
    } else {
        joinBtn.value = "JOIN";
        joinBtn.onclick = joinGame;
    }
}

function createGame() {
    gameName = document.getElementById("gameName").value;
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
        gameNameElt = document.getElementById("gameName");
        nameElt = document.getElementById("name")
        if(gameNameElt != null)
            gameName = gameNameElt.value;
        if(nameElt != null)
            name = nameElt.value;
        var messageJson = '{"gameName":"' + gameName.toUpperCase() + '", "name":"' + name.toUpperCase() + '"}';
        ws.send(messageJson);
        isconnected = true;
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
            if(rememberInv >= 0) {
                if(confirm("Are you sure you want to investigate?")) {
                    ws.send(rememberInv);
                }
                rememberInv = -1;
            } else {
                document.getElementById("beginForm").remove();
            }
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
        isconnected = false;
    };
}

function sendInvestigationMessage(index) {
    if(isconnected == false) {
        joinGame();
        rememberInv = index;
    } else if(confirm("Are you sure you want to investigate?")) {
        ws.send(index);
    }
}