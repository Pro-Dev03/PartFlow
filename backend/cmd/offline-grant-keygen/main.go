package main

import (
	"log"
)

// Offline grants are retired because a disconnected client cannot establish
// subscription state using its own mutable clock.
func main() {
	log.Fatal("offline grants are disabled; subscription authorization requires a live cloud decision")
}
