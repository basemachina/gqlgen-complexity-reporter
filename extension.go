package complexitymetrics

import (
	"context"

	"github.com/99designs/gqlgen/complexity"
	"github.com/99designs/gqlgen/graphql"
)

// ComplexityReporter is an interface for reporting a query complexity exceeds a limit
type ComplexityReporter interface {
	ReportComplexity(ctx context.Context, operationName string, complexity int)
}

// extension allows you to define a metrics reporter depending on query complexity and a defined limit.
type extension struct {
	complexityReporter ComplexityReporter
	limit              int
	es                 graphql.ExecutableSchema
}

// implements HandlerExtension
var _ graphql.HandlerExtension = (*extension)(nil)

const ExtensionName = "ComplexityReporter"

// NewComplexityReporterExtension sets a logger/tracer which reports a query complexity exceeds a limit
func NewComplexityReporterExtension(limit int, complexityReporter ComplexityReporter) *extension {
	return &extension{
		complexityReporter: complexityReporter,
		limit:              limit,
	}
}

func (c extension) ExtensionName() string {
	return ExtensionName
}

func (c *extension) Validate(schema graphql.ExecutableSchema) error {
	c.es = schema
	return nil
}

// implements OperationInterceptor
var _ graphql.OperationInterceptor = (*extension)(nil)

func (c extension) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	rc := graphql.GetOperationContext(ctx)
	if rc == nil {
		return nil
	}

	op := rc.Doc.Operations.ForName(rc.OperationName)
	complexityCalcs := complexity.Calculate(c.es, op, rc.Variables)

	if complexityCalcs > c.limit {
		c.complexityReporter.ReportComplexity(ctx, rc.OperationName, complexityCalcs)
	}

	return next(ctx)
}
