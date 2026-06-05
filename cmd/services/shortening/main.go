package main

import (
	"os"

	v1 "github.com/AdventurerAmer/shortner/cmd/services/shortening/v1"
)

func main() {
	os.Exit(v1.Run())
}
