package main

import (
	"os"

	shorteningV1 "github.com/AdventurerAmer/shortner/cmd/services/shortening/v1"
)

func main() {
	os.Exit(shorteningV1.Run())
}
