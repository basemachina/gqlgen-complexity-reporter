package complexityreporter_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	complexityreporter "github.com/basemachina/gqlgen-complexity-reporter"
	"github.com/stretchr/testify/require"

	"github.com/99designs/gqlgen/graphql/handler/testserver"
	"github.com/99designs/gqlgen/graphql/handler/transport"
)

type testReporter struct {
	complexity int
}

func (t *testReporter) ReportComplexity(ctx context.Context, _ string, complexity int) {
	t.complexity = complexity
}

func TestReportComplexity(t *testing.T) {
	h := testserver.New()
	h.AddTransport(&transport.POST{})

	t.Run("always report", func(t *testing.T) {
		reporter := &testReporter{complexity: 0}
		h.Use(complexityreporter.NewExtension(reporter))
		h.SetCalculatedComplexity(2)
		resp := doRequest(h, "POST", "/graphql", `{"query":"{ name }"}`)
		require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())

		require.Equal(t, 2, reporter.complexity)
	})

	t.Run("below complexity threshold", func(t *testing.T) {
		reporter := &testReporter{complexity: 0}
		h.Use(complexityreporter.NewExtension(reporter, complexityreporter.WithThreshold(2)))
		h.SetCalculatedComplexity(2)
		resp := doRequest(h, "POST", "/graphql", `{"query":"{ name }"}`)
		require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())

		require.Equal(t, 0, reporter.complexity)
	})

	t.Run("above complexity threshold", func(t *testing.T) {
		reporter := &testReporter{complexity: 0}
		h.Use(complexityreporter.NewExtension(reporter, complexityreporter.WithThreshold(2)))
		h.SetCalculatedComplexity(4)
		resp := doRequest(h, "POST", "/graphql", `{"query":"{ name }"}`)
		require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())

		require.Equal(t, 4, reporter.complexity)
	})

	t.Run("bypass __schema field", func(t *testing.T) {
		reporter := &testReporter{complexity: 0}
		h.Use(complexityreporter.NewExtension(reporter, complexityreporter.WithThreshold(2)))
		h.SetCalculatedComplexity(4)
		resp := doRequest(h, "POST", "/graphql", `{ "operationName":"IntrospectionQuery", "query":"query IntrospectionQuery { __schema { queryType { name } mutationType { name }}}"}`)
		require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())

		require.Equal(t, 0, reporter.complexity)
	})
}

func doRequest(handler http.Handler, method string, target string, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)
	return w
}
