$(function () {
    var loginModal = $("#loginModal");
    var logoutButton = $("#logoutBtn");
    var signupEmail = $("#signupInputEmail");
    var signupPassword = $("#signupInputPassword");
    var loginEmail = $("#loginInputEmail");
    var loginPassword = $("#loginInputPassword");

    var user;

    var accessToken = localStorage.getItem("accessToken");
    if (accessToken) {
        // Use the current access token
        logoutButton.removeClass("hidden");
        loginModal.trigger("accessToken", [accessToken, localStorage.getItem("user")]);
    } else {
        // Show the login modal
        logoutButton.addClass("hidden");
        loginModal.modal();
    }

    // Logout button handler
    logoutButton.on("click", function () {
        logoutButton.addClass("hidden");
        localStorage.removeItem("accessToken");
        window.location.reload();
        loginModal.modal();
    });

    // XHR fail handler
    function onFail(jqXHR) {
        var json = jqXHR.responseJSON;
        alert("Error: " + json.message + " " + JSON.stringify(json.details));
    }

    // XHR success handler, common to login & register
    function onSuccess(data) {
        logoutButton.removeClass("hidden");
        localStorage.setItem("accessToken", data.accessToken);
        localStorage.setItem("user", user);
        loginModal.trigger("accessToken", [data.accessToken, user]);
        loginModal.modal("hide")
    }

    // Ok button handler
    $("#loginButton").on("click", function () {
        var active = $("#loginTab .active").attr("aria-controls");
        if (active == "login") {
            login()
        } else {
            signup()
        }
    });

    // Signup
    function signup() {
        user = signupEmail.val();
        jQuery.ajax({
            url: "/register",
            type: "POST",
            data: JSON.stringify({
                email: signupEmail.val(),
                password: signupPassword.val()
            }),
            contentType: "application/json; charset=utf-8",
            dataType: "json"
        }).done(onSuccess).fail(onFail);
    }

    // Login
    function login() {
        user = loginEmail.val();
        jQuery.ajax({
            url: "/login",
            type: "POST",
            data: JSON.stringify({
                email: loginEmail.val(),
                password: loginPassword.val()
            }),
            contentType: "application/json; charset=utf-8",
            dataType: "json"
        }).done(onSuccess).fail(onFail);
    }
});
