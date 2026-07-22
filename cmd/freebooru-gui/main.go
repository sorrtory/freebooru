// Package main provides the FreeBooru graphical application.
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/sorrtory/freebooru/internal/webapi"
	"github.com/sorrtory/freebooru/internal/webui"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	assets, err := webui.Assets()
	if err != nil {
		return fmt.Errorf("load web assets: %w", err)
	}
	api := webapi.New(webapi.ModeDesktop)

	if err := wails.Run(&options.App{
		Title:     "FreeBooru",
		Width:     1080,
		Height:    720,
		MinWidth:  360,
		MinHeight: 520,
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Middleware: apiMiddleware(api),
		},
		BackgroundColour: &options.RGBA{R: 16, G: 18, B: 15, A: 1},
	}); err != nil {
		return fmt.Errorf("run Wails application: %w", err)
	}
	return nil
}

func apiMiddleware(api http.Handler) assetserver.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			if request.URL.Path == "/api" || strings.HasPrefix(request.URL.Path, "/api/") {
				api.ServeHTTP(response, request)
				return
			}
			next.ServeHTTP(response, request)
		})
	}
}
