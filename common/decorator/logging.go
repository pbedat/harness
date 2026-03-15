package decorator

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

type commandLoggingDecorator[C any] struct {
	base   CommandHandler[C]
	logger zerolog.Logger
}

var BodyField = "command_body"

type logLevelKeyType string

var LogLevelKey = logLevelKeyType("decorator.log_level")

func WithLogLevel(ctx context.Context, lvl zerolog.Level) context.Context {
	return context.WithValue(ctx, LogLevelKey, lvl)
}

func getLogLevel(ctx context.Context) (zerolog.Level, bool) {
	if ctx == nil {
		return zerolog.DebugLevel, false
	}
	if v := ctx.Value(LogLevelKey); v != nil {
		if level, ok := v.(zerolog.Level); ok {
			return level, true
		}
	}
	return zerolog.DebugLevel, false
}

func (d commandLoggingDecorator[C]) Handle(ctx context.Context, cmd C) (err error) {
	handlerType := generateActionName(cmd)

	logger := d.logger.With().
		Str(
			"command", handlerType).
		Str(BodyField, fmt.Sprintf("%#v", cmd)).
		Logger()

	if level, ok := getLogLevel(ctx); ok {
		logger = logger.Level(level)
	}

	now := time.Now()

	logger.Debug().Msg("Executing command")
	defer func() {
		p := recover()

		if p != nil {
			logger.Error().Msgf("panicked: %v", p)
			panic(p)
		}

		if err == nil {
			logger.Info().Dur("duration", time.Since(now)).Msg("Command executed successfully")
		} else {
			logger.Err(err).Msg("Failed to execute command")
		}
	}()

	err = d.base.Handle(ctx, cmd)

	return
}

type queryLoggingDecorator[C any, R any] struct {
	base   QueryHandler[C, R]
	logger zerolog.Logger
}

func (d queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	logger := d.logger.With().Fields(map[string]any{
		"query":      generateActionName(cmd),
		"query_body": fmt.Sprintf("%#v", cmd),
	}).Logger()

	if level, ok := getLogLevel(ctx); ok {
		logger = logger.Level(level)
	}

	now := time.Now()
	logger.Debug().Msg("Executing query")
	defer func() {
		if err == nil {
			logger.Debug().Dur("duration", time.Since(now)).Msg("Query executed successfully")
		} else {
			logger.Err(err).Msg("Failed to execute query")
		}
	}()

	result, err = d.base.Handle(ctx, cmd)

	return
}
