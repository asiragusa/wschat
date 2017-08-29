$(function () {
    var loginModal = $("#loginModal");
    var userListContainer = $("#userList");
    var rowTemplateEl = $("#usersRowTemplate");
    var userList = [];
    var active;
    var user;

    // Handler for click on user
    userListContainer.bind("click", function (e) {
        var el = $(e.target);
        if (!el.hasClass("_user_email")) {
            return;
        }
        if (active) {
            active.removeClass("active");
        }
        active = el;
        el.addClass("active");
        userListContainer.trigger("userSelected", [el.html()]);
    });

    // XHR fail handler
    function onFail(jqXHR) {
        var json = jqXHR.responseJSON;
        alert("Error: " + json.message + " " + JSON.stringify(json.details));
    }


    // On successful login fetch the user list
    loginModal.on("accessToken", function (e, accessToken, email) {
        user = email;

        function onSuccess(data) {
            console.log("received user list", data.items);
            userList = data.items;
            showUserList()
        }

        jQuery.ajax({
            url: "/users",
            type: "GET",
            dataType: "json",
            headers: {
                Authorization: "Bearer " + accessToken
            }
        }).done(onSuccess).fail(onFail);
    });

    // Show the user list
    function showUserList() {
        userListContainer.empty();
        userList.forEach(function (element) {
            if (element.email == user) {
                return;
            }
            var el = rowTemplateEl.clone();
            el.removeClass("hidden");
            el.find("._user_email").html(element.email);
            userListContainer.append(el);
        });
    }
});

