package main

import (
	"os"

	authv1 "github.com/AdventurerAmer/shortner/cmd/services/auth/v1"
)

func main() {
	os.Exit(authv1.Run())
}
