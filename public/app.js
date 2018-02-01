function myWebsocketStart(role)
{
	console.log(location.host);
    var ws = new WebSocket("ws://" + location.host + "/websocket");

    ws.onopen = function() {
        var name = document.getElementById("name").value;
        var json = '{"name":"' + name + '", "role":' + role + '}';
        ws.send(json);
        document.getElementById("name").remove();
        document.getElementById("libBtn").remove();
        document.getElementById("fascBtn").remove();
        document.getElementById("hitBtn").remove();
        
        var button = document.createElement("button");
        button.innerHTML = "Start!";
        button.className = "button";
        button.addEventListener ("click", function() {
        ws.send("start");
        });
        document.getElementsByTagName("body")[0].appendChild(button);
    }

    ws.onmessage = function (evt)
    {
        var personList = document.getElementById("personList");
        personList.innerHTML = (evt.data);
    };

    ws.onclose = function()
    {
        console.log("I am now closed");
    };
}