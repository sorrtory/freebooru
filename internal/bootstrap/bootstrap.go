// Package bootstrap assembles a Core for executable entry points.
package bootstrap

import (
	"log/slog"

	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/core"
)

func NewCore(logger *slog.Logger) (*core.Core, error) {
	paths, err := config.DefaultPaths()
	if err != nil {
		return nil, err
	}
	return core.New(logger, paths)
}
