$(function () {

    var connected = false;
    var from;
    var to;
    var loginModal = $("#loginModal");
    var messageTxt = $("#messageTxt");
    var messagesEl = $("#messages");
    var userListContainer = $("#userList");
    var messageListContainer = $("#messageList");
    var messageRowTpl = $("#messageRowTemplate");
    var myMessageRowTpl = $("#myMessageRowTemplate");
    var sendButton = $("#sendBtn");
    var w;

    var messages = {};

    var requestId = 0;

    // Send the message on sendButton click
    sendButton.click(sendMessage);

    // Send the message on enter key
    messageTxt.on("keyup", function (e) {
        if (e.keyCode == 13) {
            sendMessage()
        }
    });

    // When logged in
    loginModal.on("accessToken", function (e, accessToken, user) {
        connected = false;
        from = user;
        getMessages(accessToken);
        getWsToken(accessToken);
    });

    // When a user has been selected on the right panel
    userListContainer.on("userSelected", function (e, email) {
        to = email;
        messageTxt.prop("disabled", false);
        messageTxt.prop("placeholder", messageTxt.data("send-placeholder") + to);
        sendButton.prop("disabled", false);
        showMessages();
    });

    // XHR fail handler
    function onFail(jqXHR) {
        var json = jqXHR.responseJSON;
        alert("Error: " + json.message + " " + JSON.stringify(json.details));
    }

    // Send a message
    function sendMessage() {
        if (!connected) {
            return
        }
        var message = messageTxt.val();
        w.Emit("message", {
            "requestId": "" + requestId,
            "body": {
                "to": to,
                "message": message
            }
        });
        messages[requestId] = {
            from: from,
            to: to,
            message: message,
            createdAt: new Date(),
            sending: true
        };
        requestId++;
        messageTxt.val("");
        showMessages();
    }

    // Adds a message received from the server to the messages object
    function addMessage(item) {
        messages[item.id] = {
            from: item.from,
            to: item.to,
            message: item.message,
            createdAt: new Date(item.createdAt)
        }
    }

    // Retrieve a valid token from the server to connect to the websocket
    function getWsToken(accessToken) {
        function onSuccess(data, textStatus, jqXHR) {
            connect(data.token)
        }

        jQuery.ajax({
            url: "/wsToken",
            type: "POST",
            dataType: "json",
            headers: {
                Authorization: "Bearer " + accessToken
            }
        }).done(onSuccess).fail(onFail);
    }

    // Fetches the existing messages from the server
    function getMessages(accessToken) {
        function onSuccess(data) {
            messages = {};
            data.items.forEach(addMessage);
            showMessages();
        }

        jQuery.ajax({
            url: "/messages",
            type: "GET",
            dataType: "json",
            headers: {
                Authorization: "Bearer " + accessToken
            }
        }).done(onSuccess).fail(onFail);
    }

    // Connects to the websocket
    function connect(wsToken) {
        w = new Ws("ws://" + window.location.host + "/ws?token=" + wsToken);

        w.OnConnect(function () {
            connected = true;
            console.log("Websocket connection enstablished");
        });

        w.OnDisconnect(function () {
            connected = false;
            alert('disconnected');
            window.location.reload();
        });
        w.On("sent", function (data) {
            console.log("Message successfully sent", data);
            var decoded = JSON.parse(data);
            if (!decoded || !decoded.body) {
                return;
            }
            delete (messages[decoded.requestId]);
            addMessage(decoded.body);
            showMessages();
        });
        w.On("error", function (data) {
            console.error("error", data);

            var decoded = JSON.parse(data);
            if (!decoded || !decoded.requestId) {
                return;
            }
            messages[decoded.requestId].sending = false;
            messages[decoded.requestId].error = true;
            showMessages();
        });
        w.On("message", function (data) {
            console.log("Received message", data);
            var decoded = JSON.parse(data);
            if (!decoded || !decoded.body) {
                return;
            }
            addMessage(decoded.body);
            showMessages();
        });
    }

    // Shows the messages
    function showMessages() {
        var keys = [];
        for (var id in messages) {
            m = messages[id]
            if (m.from == to || m.to == to) {
                keys.push(id);
            }
        }
        keys.sort(function (a, b) {
            return messages[a].createdAt - messages[b].createdAt;
        });

        messageListContainer.empty();
        keys.forEach(function (id) {
            var message = messages[id];
            var tpl = messageRowTpl;
            if (message.from == from) {
                tpl = myMessageRowTpl;
            }
            var el = tpl.clone();
            el.removeClass("hidden");
            el.find("._message").html(message.message);
            el.find("._date").html(message.createdAt.getHours() + ":" + message.createdAt.getMinutes());
            if (message.sending) {
                el.addClass("sending");
            }
            if (message.error) {
                el.addClass("error");
                el.find("._has_error").removeClass("hidden")
            }
            messageListContainer.append(el);
        });

        messagesEl.scrollTop(messageListContainer.height());
    }
});

