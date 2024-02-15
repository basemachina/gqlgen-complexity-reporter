package complexitymetrics

import (
	"context"

	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/99designs/gqlgen/complexity"
	"github.com/99designs/gqlgen/graphql"
)

// ComplexityMetrics allows you to define a metrics reporter depending on query complexity and a defined limit.
type ComplexityMetrics struct {
	reporter Reporter
	limit    int
	es       graphql.ExecutableSchema
}

type Reporter interface {
	Report(ctx context.Context, operationName string, complexity int)
}

var _ interface {
	graphql.OperationContextMutator
	graphql.HandlerExtension
} = &ComplexityMetrics{}

const ExtensionName = "ComplexityMetrics"

type ComplexityStats struct {
	// The calculated complexity for this request
	Complexity int
}

// ReportComplexity sets a logger/tracer which reports a query complexity exceeds a limit
func ReportComplexity(limit int, reporter Reporter) *ComplexityMetrics {
	return &ComplexityMetrics{
		reporter: reporter,
		limit:    limit,
	}
}

func (c ComplexityMetrics) ExtensionName() string {
	return ExtensionName
}

func (c *ComplexityMetrics) Validate(schema graphql.ExecutableSchema) error {
	c.es = schema
	return nil
}

func (c ComplexityMetrics) MutateOperationContext(ctx context.Context, rc *graphql.OperationContext) *gqlerror.Error {
	op := rc.Doc.Operations.ForName(rc.OperationName)
	complexityCalcs := complexity.Calculate(c.es, op, rc.Variables)

	rc.Stats.SetExtension(ExtensionName, &ComplexityStats{
		Complexity: complexityCalcs,
	})

	if complexityCalcs > c.limit {
		c.reporter.Report(ctx, rc.OperationName, complexityCalcs)
	}

	return nil
}
