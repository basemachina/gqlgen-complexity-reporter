package complexitymetrics_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	complexitymetrics "github.com/basemachina/gqlgen-complexity-metrics"
	"github.com/stretchr/testify/require"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/testserver"
	"github.com/99designs/gqlgen/graphql/handler/transport"
)

type testReporter struct {
	complexity int
}

func (t *testReporter) Report(ctx context.Context, _ string, complexity int) {
	t.complexity = complexity
}

func getComplexityStats(ctx context.Context) *complexitymetrics.ComplexityStats {
	rc := graphql.GetOperationContext(ctx)
	if rc == nil {
		return nil
	}

	s, _ := rc.Stats.GetExtension(complexitymetrics.ExtensionName).(*complexitymetrics.ComplexityStats)
	return s
}

func TestReportComplexity(t *testing.T) {
	h := testserver.New()
	h.AddTransport(&transport.POST{})

	var stats *complexitymetrics.ComplexityStats
	h.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		stats = getComplexityStats(ctx)
		return next(ctx)
	})

	t.Run("below complexity limit", func(t *testing.T) {
		reporter := &testReporter{complexity: 0}
		h.Use(complexitymetrics.ReportComplexity(2, reporter))
		h.SetCalculatedComplexity(2)
		resp := doRequest(h, "POST", "/graphql", `{"query":"{ name }"}`)
		require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())

		require.Equal(t, 2, stats.Complexity)
		require.Equal(t, 0, reporter.complexity)
	})

	t.Run("above complexity limit", func(t *testing.T) {
		reporter := &testReporter{complexity: 0}
		h.Use(complexitymetrics.ReportComplexity(2, reporter))
		h.SetCalculatedComplexity(4)
		resp := doRequest(h, "POST", "/graphql", `{"query":"{ name }"}`)
		require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())

		require.Equal(t, 4, stats.Complexity)
		require.Equal(t, 4, reporter.complexity)
	})

	t.Run("bypass __schema field", func(t *testing.T) {
		reporter := &testReporter{complexity: 0}
		h.Use(complexitymetrics.ReportComplexity(2, reporter))
		h.SetCalculatedComplexity(4)
		resp := doRequest(h, "POST", "/graphql", `{ "operationName":"IntrospectionQuery", "query":"query IntrospectionQuery { __schema { queryType { name } mutationType { name }}}"}`)
		require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())

		require.Equal(t, 0, stats.Complexity)
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
