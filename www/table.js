(function() {
    $(document).ready(function() {
        $("table th").click(function(e) {
            localStorage.setItem("table_sort", Array.from(e.target.parentElement.children).indexOf(e.target).toString());
        });
        const ts = localStorage.getItem("table_sort");
        if (ts !== null && ts.length > 0) {
            document.querySelectorAll("table th")[parseInt(ts)].setAttribute("data-sort-default", "");
        }
        new Tablesort(document.querySelector("table"), {
            descending: false,
        });
        const mel = document.querySelector(".modal .content").children;
        $("table tr td i.info.icon").click(function(e) {
            const f = e.target.parentElement.parentElement.children[2].firstElementChild.pathname;
            const url = "/api/search?q="+f;
            fetch(url).then(x => x.json()).then(x => {
                console.log('we got funny');
                if (x.count === 0) {
                    console.log('unfunny');
                    return
                }
                mel[0].parentElement.parentElement.children[1].textContent = x.results[0].path;
                mel[0].children[0].children[1].value = x.results[0].hash_md5;
                mel[1].children[0].children[1].value = x.results[0].hash_sha1;
                $(".ui.basic.modal").modal("show");
            })
        });
    })
})();