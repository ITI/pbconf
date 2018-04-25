function onLogin(){
    uname = $(".form-signin input[name=User]");
    upass = $(".form-signin input[name=Pass]");
    dd = '{"' + $(uname).attr("name") + '":"' + $(uname).val() + '",' +
    '"' + $(upass).attr("name") + '":"' + $(upass).val() +
    '"}';
    $.ajax({
        url:"/ui/login",
        method:"POST",
        data:dd,
    }).done(function(){
        $.ajax({
            url:"/ui",
            method:"GET",
        }).done(function(newContent){
            document.open();
            document.write(newContent);
            document.close();
        });
    }).fail(function(){
        alert("Login Failed!! Try again.");
    });
}

$("#pass").keyup(function(event){
    if(event.keyCode == 13){
        $("#login_btn").click();
    }
});
