package main

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/zaz600/go-musthave-diploma/internal/app"
)

func main() {
	os.Exit(CLI(os.Args))
}

func CLI(args []string) int {
	if err := app.Run(args); err != nil {
		log.Err(err).Msgf("Runtime error")
		return 1
	}
	return 0
}
