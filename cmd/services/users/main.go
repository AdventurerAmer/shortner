package main

import (
	"os"

	redirectingV1 "github.com/AdventurerAmer/shortner/cmd/services/redirecting/v1"
)

func main() {
	os.Exit(redirectingV1.Run())
}
