jQuery(document).ready(function() {
    var shown=false;
    $("tbody").each(function () {$(this).hide()});
    $("#show_all").click(function(e){
        $("tbody").each(function() {
            if (shown)
                $(this).hide();
            else
                $(this).show();
        });
        shown=!shown;
        e.preventDefault();
    });

    $(".repository_name").click(function(e){
        $(this).next("tbody").toggle();
        e.preventDefault();
    });

});
