# Scrape je stad

A Go library for scraping data from the
[Meet je stad](https://meetjestad.net)
project.

_This is not affiliated with Meet je stad in any way_

## Usage

Example program:

```go
package main

import (
    "github.com/fiskeben/scrapejestad"
    "net/url"
)

func main() {
    u, err := url.Parse("https://meetjestad.net/data/sensors_recent.php?sensor=242&limit=10")
    if err != nil {
        panic(err)
    }

    data, err := scrapejestad.Read(u)
    if err != nil {
        panic(err)
    }

    // do stuff to data
    fmt.Printf("first entry: %v\n", data[0])
}
```

## See also

See the
[meetjescraper](https://github.com/fiskeben/meetjescraper)
HTTP proxy to get a JSON based HTTP API in front of Meet je stad.
