# osrswiki

A library for interacting with the Old School Runescape Wiki API.

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joeychilson/osrswiki"
)

func main() {
	client := osrswiki.NewClient("osrswiki-cli/0.0.1")

	data, err := client.Timeseries(context.Background(), osrswiki.WorldRegular, osrswiki.FiveMinutes, 4151)
	if err != nil {
		log.Fatalf("error getting latest prices: %v", err)
	}

	fmt.Println(data)
}
```
