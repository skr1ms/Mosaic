// MIT License
//
// Copyright (c) 2025 Mosaic Project
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/skr1ms/mosaic/pkg/middleware"
)

var mainLogger = middleware.NewLogger()

func main() {
	app := InitializeApp()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		mainLogger.GetZerologLogger().Info().Msg("Starting server on port 8080")
		app.Listen(":8080")
	}()

	select {
	case <-sigChan:
		mainLogger.GetZerologLogger().Info().Msg("Shutdown signal received")
	case <-ctx.Done():
		mainLogger.GetZerologLogger().Info().Msg("Context cancelled")
	}

	mainLogger.GetZerologLogger().Info().Msg("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		mainLogger.GetZerologLogger().Error().Err(err).Msg("Server forced to shutdown")
	} else {
		mainLogger.GetZerologLogger().Info().Msg("Server gracefully stopped")
	}

	mainLogger.GetZerologLogger().Info().Msg("Application shutdown complete")
}
