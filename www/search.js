$(document).ready(function() {
    const base = "/".slice(0, -1);
    $(".ui.search").search({
        apiSettings: {
            url: "/api/search?q={query}&w="+window.location.pathname,
            onResponse: function(api_response) {
                return {
                    results: api_response.results.map(function(v) {
                        return {
                            title: v.path,
                            description: v.html_modtime + " || " + v.html_size,
                            url: base + v.path,
                        };
                    }),
                }
            },
        },
        maxResults: 25,
        showNoResults: true,
    });
});
