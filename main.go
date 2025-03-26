package main

import (
	"fmt"
	"images-of-the-day/cmd"
	"os"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
