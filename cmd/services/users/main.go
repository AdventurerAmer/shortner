package main

import (
	"os"

	usersv1 "github.com/AdventurerAmer/shortner/cmd/services/users/v1"
)

func main() {
	os.Exit(usersv1.Run())
}
