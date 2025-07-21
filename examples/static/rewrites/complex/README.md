### examples/static/rewrites/complex

This example demonstrates how dynamic rewrites with prefixes and suffixes can be used.
Run the server using `main.go`.

The website is now accessible at [`localhost:8080`](http://localhost:8080/) and features a number of pages, images (of different mime types) and stylesheets.
The content (including images) was generated using ChatGPT (specifically `gpt-4-0125-preview`).

The rewrite rules can be found in [`config/rewrites.conf`](./config/rewrites.conf).
There are 5 rules, which are explained in the file.

We have 1 static rewrite: `/ /static/html/index.html`.

Then there are four dynamic rewrites, three of which use required suffixes.

`/${style||.css} /static/css/${style}`, `/${img||.png} /static/images/${img}` and `/${img||.jpg} /static/images/${img}` all rewrite appropriate stylesheet and image content to the corresponding static directory.
Notice how there is no collision here, despite having three dynamic matches that match against the same path component.
This is because all three require a separate suffix, so the set of routes that match against them are all completely separated.

Lastly, we have `/${page} /static/html/${page}.html`.
Although this will match against any route that also matches against the three rules above, it is unambiguously less specific than the three rules above, so it a route also matches one of those three, it will always choose that rewrite instead of this one.
This one allows us to omit the `.html` suffix from the URLs that our users see in the browser, which improves the experience and makes mistakes less likely to occur.
