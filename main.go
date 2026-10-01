// ABOUTME: Entrypoint for the intervals.icu MCP server. Loads config, wires the HTTP client,
// ABOUTME: registers all domain tools (activities, wellness, events, athlete) and serves over stdio.

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/mark3labs/mcp-go/server"

	"github.com/maroffo/intervals-icu-mcp/internal/config"
	"github.com/maroffo/intervals-icu-mcp/internal/icu"
	"github.com/maroffo/intervals-icu-mcp/internal/tools"
)

const (
	serverName = "intervals-icu-mcp"
	logPrefix  = "intervals-icu-mcp: "
)

// serverVersion is overridable at build time via:
//
//	go build -ldflags "-X main.serverVersion=1.2.3"
//
// When unset it falls back to build info (vcs.revision) or "dev".
var serverVersion = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(ctx, logger); err != nil {
		// Match the "intervals-icu-mcp: <msg>" convention so operators can
		// visually attribute the failure to this process in aggregated logs.
		fmt.Fprintln(os.Stderr, logPrefix+err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		if errors.Is(err, config.ErrMissingAPIKey) {
			return errors.New(`INTERVALS_API_KEY is required. Set it in the "env" block of your MCP client config (see README)`)
		}
		return err
	}

	version := resolveVersion()
	logger.Info("starting intervals-icu-mcp",
		slog.String("name", serverName),
		slog.String("version", version),
		slog.String("base_url", cfg.BaseURL),
		slog.String("athlete_id", maskAthleteID(cfg.AthleteID)),
	)

	client := icu.New(cfg.BaseURL, cfg.APIKey, cfg.AthleteID)

	s := server.NewMCPServer(serverName, version,
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	tools.RegisterActivities(s, client)
	tools.RegisterWellness(s, client)
	tools.RegisterEvents(s, client)
	tools.RegisterAthlete(s, client)

	// ServeStdio blocks; when ctx is canceled (SIGINT/SIGTERM) the underlying
	// stdio reader is closed by the signal handler's process termination path.
	// The goroutine below ensures a log line is emitted on shutdown signal.
	go func() {
		<-ctx.Done()
		logger.Info("shutdown signal received", slog.String("cause", context.Cause(ctx).Error()))
	}()

	if err := server.ServeStdio(s); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
	return nil
}

// resolveVersion prefers an ldflags-injected serverVersion, then falls back to
// the VCS revision embedded by `go build`, and finally to "dev".
func resolveVersion() string {
	if serverVersion != "" && serverVersion != "dev" {
		return serverVersion
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" && s.Value != "" {
				rev := s.Value
				if len(rev) > 12 {
					rev = rev[:12]
				}
				return "dev+" + rev
			}
		}
	}
	return serverVersion
}

// maskAthleteID returns a log-safe representation: first 2 chars + "***".
// Empty or very short IDs are replaced entirely with "***".
func maskAthleteID(id string) string {
	if len(id) < 3 {
		return "***"
	}
	return id[:2] + "***"
}
