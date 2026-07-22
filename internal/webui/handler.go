package webui

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// NewHandler serves an API alongside a Vue single-page application.
func NewHandler(api http.Handler, assets fs.FS) http.Handler {
	files := http.FileServer(http.FS(assets))
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, "/api/") {
			api.ServeHTTP(response, request)
			return
		}
		if request.Method != http.MethodGet && request.Method != http.MethodHead {
			response.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		requested := strings.TrimPrefix(path.Clean(request.URL.Path), "/")
		if requested == "." || requested == "" {
			requested = "index.html"
		}
		if info, err := fs.Stat(assets, requested); err == nil && !info.IsDir() {
			files.ServeHTTP(response, request)
			return
		}

		indexRequest := request.Clone(request.Context())
		indexRequest.URL.Path = "/"
		files.ServeHTTP(response, indexRequest)
	})
}
