package chat

// Html
var HTML = `
<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Document</title>
    <style type="text/css">
        html,
		body,
		iframe {
			height: 100%;
			margin: 0;
        }

        .container {
            height: 100%;
        }

        iframe {
            width: 100%;
            height: 98%;
			border: 0;
			box-sizing: border-box;
        }
    </style>
</head>

<body style="margin: 0;">
    <div class="container">
        <iframe id="frame" src=""></iframe>
    </div>

    <script>
        var uri = "ws://127.0.0.1:8080/entry";
        console.log("connect to " + uri);
        var socket = new WebSocket(uri);
        socket.onopen = function () {
            console.log("connected to " + uri);
        }
        socket.onclose = function (e) {
            console.log("connection closed (" + e.code + ")");
        }
        socket.onmessage = function (e) {
            console.log("message received: " + e.data);
			var frame = document.getElementById("frame");
			console.log("scroll", frame.contentWindow.scrollTop);
            var replaceTo = "data" + JSON.parse(e.data).path;
            frame.contentWindow.location.replace(replaceTo);
        }
    </script>
</body>

</html>
`
