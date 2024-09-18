package middleware

import (
	"net/http"
	"path"
)

func AllowPath(allowedPaths []string, next http.Handler) http.Handler {
	allowedPathsSet := map[string]interface{}{}

	for _, path := range allowedPaths {
		allowedPathsSet[path] = struct{}{}
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if _, ok := allowedPathsSet[path.Clean(r.URL.Path)]; !ok {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			next.ServeHTTP(w, r)
		},
	)
}
