var ws;

function myWebsocketStart()
{
    ws = new WebSocket("ws://" + location.host + "/websocket");

    ws.onopen = function() {
        var name = document.getElementById("name").value;
        ws.send(name.toUpperCase());
        
    }

    ws.onmessage = function (evt)
    {
        if(evt.data == "Not enough players" || 
        evt.data == "Game full" ||
        evt.data == "Game started" ||
        evt.data == "Duplicate name") {
            alert(evt.data);
        } else if (evt.data == "Joined"){
            document.getElementById("name").remove();
            document.getElementById("enterBtn").remove();
        
            var button = document.createElement("button");
            button.innerHTML = "START";
            button.className = "button";
            button.addEventListener ("click", function() {
            ws.send("start");
            });
            document.getElementsByTagName("body")[0].appendChild(button);
        } else {
            var personList = document.getElementById("personList");
            personList.innerHTML = (evt.data);
        }
    };

    ws.onclose = function()
    {
        console.log("I am now closed");
    };
}

function sendInvestigationMessage(index) {
    ws.send(index);
}