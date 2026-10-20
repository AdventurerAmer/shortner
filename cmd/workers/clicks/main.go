package main

import (
	"os"

	clicksv1 "github.com/AdventurerAmer/shortner/cmd/workers/clicks/v1"
)

func main() {
	os.Exit(clicksv1.Run())
}
