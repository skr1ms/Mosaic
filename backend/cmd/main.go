package main

import "github.com/rs/zerolog/log"

func main() {
	app := InitializeApp()

	log.Info().Msg("Starting server on port 8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
