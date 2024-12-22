package main

import (
	"check-the-score/nfl"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: check-the-score <api-name>")
		fmt.Println("Supported APIs: nfl")
		os.Exit(1)
	}

	apiName := os.Args[1]

	switch apiName {
	case "nfl":
		nfl.DisplayNFLGames()
	default:
		fmt.Printf("Unsupported API: %s\n", apiName)
		fmt.Println("Supported APIs: nfl")
		os.Exit(1)
	}
}
