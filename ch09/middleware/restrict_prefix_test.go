package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRestrictPrefix(t *testing.T) {
	handler := http.StripPrefix("/static/", RestrictPrefix(".", http.FileServer(http.Dir("./files/"))))

	testCases := []struct {
		path string
		code int
	}{
		{"http://test/static/sage.svg", http.StatusOK},
		{"http://test/static/.secret", http.StatusNotFound},
		{"http://test/static/.dir/secret", http.StatusNotFound},
	}

	for i, c := range testCases {
		r := httptest.NewRequest(http.MethodGet, c.path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		actual := w.Result().StatusCode
		if c.code != actual {
			t.Errorf("%d: expected %d; actual %d", i, c.code, actual)
		}
	}
}

func TestAllowPath(t *testing.T) {
	handler := http.StripPrefix("/static/", AllowPath([]string{".secret", ".dir/secret"}, http.FileServer(http.Dir("./files/"))))

	testCases := []struct {
		path string
		code int
	}{
		{"http://test/static/.secret", http.StatusOK},
		{"http://test/static/.dir/secret", http.StatusOK},
		{"http://test/static/this_file_dne.svg", http.StatusNotFound},
		{"http://test/static/sage.svg", http.StatusNotFound},
	}

	for i, c := range testCases {
		r := httptest.NewRequest(http.MethodGet, c.path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		actual := w.Result().StatusCode
		if c.code != actual {
			t.Errorf("%d: expected %d; actual %d", i, c.code, actual)
		}
	}
}
