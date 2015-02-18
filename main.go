package main

import (
	"fmt"
	"time"
)

func main() {
	q := NewQuery("chromium")
	q = q.Label("cr-ui-input-virtualkeyboard")
	q = q.OpenedBefore(time.Now().Add(-24 * time.Hour))
	fmt.Println(q.URL())
}
