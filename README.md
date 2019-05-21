## Intro

**cloudflare-bypasser**

Inspired by python module [cloudflare-scrape](https://github.com/Anorov/cloudflare-scrape)

## Install

`go get github.com/AurevoirXavier/cloudflare-bypasser-go`

## Require

- **Node.js**

## Example

```go
package main

import (
    "fmt"
    bypasser "github.com/AurevoirXavier/cloudflare-bypasser-go"
    "net/http"
    "os"
)

func main() {
    var (
        _            = os.Setenv("HTTP_PROXY", "http://127.0.0.1:1087")
        client       = bypasser.NewBypasser(http.DefaultClient)
        req, _       = http.NewRequest("GET", "http://cosplayjav.pl", nil)
        u, cookie, _ = client.Solve(req, 0)
    )

    fmt.Println(u)
    fmt.Println(cookie)
}
```
