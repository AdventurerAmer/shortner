package main

import (
	"os"

	analyticsV1 "github.com/AdventurerAmer/shortner/cmd/services/analytics/v1"
)

func main() {
	os.Exit(analyticsV1.Run())
}
