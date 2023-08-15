package loader_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/mxmauro/configreader"
	"github.com/mxmauro/configreader/internal/testhelpers"
	"github.com/mxmauro/configreader/loader"
)

//------------------------------------------------------------------------------

func TestHttpLoader(t *testing.T) {
	// Create a test http server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route
		switch r.Method {
		case "GET":
			switch r.URL.Path {
			case "/settings":
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(testhelpers.GoodSettingsJSON))
				return
			}
		}

		// Else return bad request
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	// Load configuration from web
	settings, err := configreader.New[testhelpers.TestSettings]().
		WithLoader(loader.NewHttp().WithHost(server.Listener.Addr().String()).WithPath("/settings")).
		WithSchema(testhelpers.SchemaJSON).
		Load(context.Background())
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Check if settings are the expected
	if !reflect.DeepEqual(settings, &testhelpers.GoodSettings) {
		t.Fatalf("settings mismatch")
	}
}
