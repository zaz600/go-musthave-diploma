package main

import (
	"log"
	"os"

	"github.com/zaz600/go-musthave-diploma/internal/app"
)

func main() {
	os.Exit(CLI(os.Args))
}

func CLI(args []string) int {
	if err := app.Run(args); err != nil {
		// log.Error().Err(err).Msgf("Runtime error")
		log.Println(err)
		return 1
	}
	return 0
}
