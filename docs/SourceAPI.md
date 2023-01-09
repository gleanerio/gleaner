## Using a paged API endpoint as a Gleaner source

Sometimes, instead of crawling webpages using a list in a sitemap, we have the opportunity to query an API that will let us directly ingest JSON-LD. To do so, we can specify a `sourcetype: api` in our Gleaner config yaml, and Gleaner will iterate through a paged API, using the given `url` as a template. For example, let's say that you want to use the API endpoint at `http://test-api.com`, and that you can page through it by using a url like `http://test/api.com/page/4`. You would put this in your config:

```yaml
url: http://test-api.com/page/%d
```

Notice the `%d` where the page number goes. Gleaner will then increment that number (starting from 0) until it gets an error back from the API.

Optionally, you can set a limit on the number of pages to iterate through, using `apipagelimit`. This means that Gleaner will page through the API until it gets an error back *or* until it reaches the limit you set. That looks like the example below:

```yaml
url: http://test-api.com/page/%d
apipagelimit: 200
```
