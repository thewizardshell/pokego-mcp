// Package main provides the entry point for the Pokemon MCP server.
// This is a learning project that implements a Model Context Protocol (MCP) server
// with Pokemon-related tools using the PokeAPI for data retrieval.
//
// The server supports two transport modes:
//   - STDIO: For Claude Desktop integration
//   - SSE: For web-based clients using Server-Sent Events
//
// Usage:
//
//	-mode string
//	    Transport mode: stdio or sse (default "sse")
//	-port int
//	    Port for SSE mode (default 8080)
//
// Examples:
//
//	./mcp-pokemon -mode=stdio              # For Claude Desktop
//	./mcp-pokemon -mode=sse -port=8080     # For web clients
//
//	d:
package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"mcp-pokemon/internal/server"
	"mcp-pokemon/internal/tools"
)

func main() {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		FormatLevel: func(i interface{}) string {
			switch i.(string) {
			case "info":
				return color.New(color.FgCyan, color.Bold).Sprint("INFO")
			case "warn":
				return color.New(color.FgYellow, color.Bold).Sprint("WARN")
			case "error":
				return color.New(color.FgRed, color.Bold).Sprint("ERROR")
			case "fatal":
				return color.New(color.FgRed, color.Bold, color.BgWhite).Sprint("FATAL")
			default:
				return color.New(color.FgBlue, color.Bold).Sprint("DEBUG")
			}
		},
		FormatMessage: func(i interface{}) string {
			return color.New(color.FgWhite).Sprint(i)
		},
	}
	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	mode := flag.String("mode", "sse", "Transport mode: stdio or sse")
	port := flag.Int("port", 8080, "Port for SSE mode")
	flag.Parse()

	pokemonBanner := color.New(color.FgYellow, color.Bold)
	pokemonBanner.Println("╔══════════════════════════════════════╗")
	pokemonBanner.Println("║          🎮 POKEMON MCP SERVER       ║")
	pokemonBanner.Println("║         Gotta catch 'em all!         ║")
	pokemonBanner.Println("╚══════════════════════════════════════╝")

	log.Info().Str("name", "pokemon-mcp").Str("version", "1.0.0").Msg("🔧 Creating MCP server")
	mcpServer := server.NewMCPServer("pokemon-mcp", "1.0.0")

	log.Info().Msg("⚙️  Setting up Pokemon tools...")
	if err := tools.SetupTools(mcpServer); err != nil {
		log.Fatal().Err(err).Msg("💥 Failed to setup tools")
	}

	log.Info().Msg("✅ Pokemon tools configured successfully")

	ctx := context.Background()

	switch *mode {
	case "sse":
		log.Info().Str("mode", "SSE").Msg("🚀 Starting Pokemon MCP Server")
		log.Info().Int("port", *port).Msg("📡 Server listening on port")

		endpoints := color.New(color.FgMagenta, color.Bold)
		endpoints.Println("🔗 Available endpoints:")
		endpoints.Printf("   • SSE Stream: http://localhost:%d/sse\n", *port)
		endpoints.Printf("   • Messages:   http://localhost:%d/message\n", *port)

		ready := color.New(color.FgGreen, color.Bold)
		ready.Println("🎮 Ready to catch 'em all!")
		ready.Println("💫 Server is running... Press Ctrl+C to stop")

		if err := mcpServer.StartSSE(ctx, *port); err != nil {
			log.Fatal().Err(err).Str("mode", "SSE").Msg("💥 Failed to start server")
		}

	case "stdio":
		log.Info().Str("mode", "STDIO").Msg("🚀 Starting Pokemon MCP Server")
		log.Info().Msg("🖥️  Claude Desktop integration mode activated")

		ready := color.New(color.FgGreen, color.Bold)
		ready.Println("🎮 Pokédex tools loaded and ready!")
		ready.Println("📡 Waiting for Claude Desktop connection...")

		if err := mcpServer.StartSdio(ctx); err != nil {
			log.Fatal().Err(err).Str("mode", "STDIO").Msg("💥 Failed to start server")
		}

	default:
		log.Fatal().Str("mode", *mode).Msg("❌ Invalid mode. Available: 'stdio' or 'sse'")
	}
}
