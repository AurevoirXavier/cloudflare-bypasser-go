## Intro

**cloudflare-bypasser**

Inspired by python module [cloudflare-scrape](https://github.com/Anorov/cloudflare-scrape)

## Require

- **Node.js**

## Example

```go
package main

import (
	bypasser "cloudflare-bypasser"
	"fmt"
	"net/http"
	"os"
)

func main() {
	var (
		client       = bypasser.NewBypasser(http.DefaultClient)
		req, _       = http.NewRequest("GET", "http://cosplayjav.pl", nil)
		u, cookie, _ = client.Solve(req, 0) // retry times, it might be 10000, depends on your network environment, 0 means infinity
	)
	
	fmt.Println(u)
	fmt.Println(cookie)
}
```
