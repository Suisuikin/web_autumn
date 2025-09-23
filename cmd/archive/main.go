package main

import (
	"math/rand"
	"rip/internal/api"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	api.StartServer()
}
