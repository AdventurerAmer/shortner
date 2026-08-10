package main

import (
	"os"

	ClicksV1 "github.com/AdventurerAmer/shortner/cmd/workers/clicks/v1"
)

func main() {
	os.Exit(ClicksV1.Run())
}
