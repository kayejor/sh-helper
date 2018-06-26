var ws;
var name = "";
var gameName = "";
var players;

function toggleCreateMode() {
    var createMode = document.getElementById("createSwitch").checked;
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
            name = nameElt.value.toUpperCase();
        var messageJson = JSON.stringify({gameName: gameName.toUpperCase(), name: name.toUpperCase()});
        console.log(messageJson);
        ws.send(messageJson);
        isconnected = true;
    }

    ws.onmessage = function (evt)
    {
        var msg = JSON.parse(evt.data);
        var type = msg["type"];
        if(type == "error") {
            alert(msg["message"]);
        } else if(type == "control") {
            if(msg["message"] == "Joined") {
                document.getElementById("beginForm").remove();
            } else if(msg["message"] == "Start") {
                var startButton = document.getElementById("startBtn");
                if(startButton != null) startButton.remove();
                createRevealButton();
            }
        } else if(type == "playerInfo") {
            players = msg["players"];
            setPlayerList();
            if(getMe() == 0) {
                createStartButton();
            }
        } else {
            //ignore
        }
    };

    ws.onclose = function()
    {
        isconnected = false;
    };
}

function setPlayerList() {
    var list = document.getElementById("personList");
    while(list.firstChild) list.removeChild(list.firstChild);
    for(var i = 0; i < players.length; i++) {
        var item = document.createElement('li');
        item.addEventListener("click", invFunc(i));
        var logoDiv = document.createElement('div');
        logoDiv.className += "logo ";
        item.appendChild(logoDiv);
        var nameDiv = document.createElement('div');
        nameDiv.className += "listName";
        nameDiv.innerText = players[i].name;
        item.appendChild(nameDiv);
        list.appendChild(item);
    }
}

function invFunc(index) {
    return function() {
        investigate(index);
    };
}

function createStartButton() {
    if(document.getElementById("startBtn") == null) {
        createButton("START", function() { ws.send("start"); });
    }
}

function createRevealButton() {
    createButton("REVEAL", revealRoles);
}

function createHideButton() {
    createButton("HIDE", hideRoles);
}

function createButton(text, onclickFunc) {
    var button = document.createElement("button");
    button.innerHTML = text;
    button.className = "button";
    button.id = text.toLowerCase() + "Btn";
    button.addEventListener ("click", onclickFunc);
    document.getElementsByTagName("body")[0].appendChild(button);
}

function revealRoles() {
    var listItems = document.getElementById("personList").getElementsByTagName("li");
    var me = getMe();
    var role = players[me]["role"];
    var amIAllKnowing = (role == "fascist") || (role == "hitler" && players.length < 7);
    if(amIAllKnowing) {
        for(var i = 0; i < listItems.length; i++) {
            reveal(listItems[i], players[i]["role"]);
        }
    } else {
        reveal(listItems[me], role);
    }
    document.getElementById("revealBtn").remove();
    createHideButton();
}

function hideRoles() {
    var listItems = document.getElementById("personList").getElementsByTagName("li");
    for(var i = 0; i < listItems.length; i++) {
        hide(listItems[i]);
    }
    document.getElementById("hideBtn").remove();
    createRevealButton();
}

function getParty(role) {
    return (role == "hitler" ? "fascist" : role);
}

function revealInv(index) {
    var listItems = document.getElementById("personList").getElementsByTagName("li");
    var role = getParty(players[index]["role"]);
    reveal(listItems[index], role);
}

function reveal(item, role) {
    console.log(item);
    var party = getParty(role);
    item.className = role;
    item.firstChild.className = "logo " + party;
}

function hide(item) {
    item.className = "";
    item.firstChild.className = "logo ";
}

function getMe() {
    var i = 0;
    while(players[i]["name"] != name) i++;
    return i;
}

function investigate(index) {
    if(confirm("Are you sure you want to investigate?")) {
        revealInv(index);
    }
}