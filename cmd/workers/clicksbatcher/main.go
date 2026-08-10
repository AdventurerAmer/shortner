package main

import (
	"os"

	clicksBatcherV1 "github.com/AdventurerAmer/shortner/cmd/workers/clicksbatcher/v1"
)

func main() {
	os.Exit(clicksBatcherV1.Run())
}
