package main

import (
	"fmt"
	"os"

	"github.com/new-er/images-of-the-day/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
