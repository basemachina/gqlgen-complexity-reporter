package complexityreporter

import (
	"context"

	"github.com/99designs/gqlgen/complexity"
	"github.com/99designs/gqlgen/graphql"
)

// ComplexityReporter is an interface for reporting a query complexity exceeds a threshold
type ComplexityReporter interface {
	ReportComplexity(ctx context.Context, operationName string, complexity int)
}

// extension allows you to define a metrics reporter depending on query complexity and a defined threshold.
type extension struct {
	complexityReporter ComplexityReporter
	threshold          int
	es                 graphql.ExecutableSchema
}

// implements HandlerExtension
var _ graphql.HandlerExtension = (*extension)(nil)

// NewExtension sets a logger/tracer which reports a query complexity exceeds a threshold
func NewExtension(complexityReporter ComplexityReporter, opts ...Option) *extension {
	e := &extension{
		complexityReporter: complexityReporter,
		threshold:          0,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

type Option func(*extension)

func WithThreshold(threshold int) func(*extension) {
	return func(e *extension) {
		e.threshold = threshold
	}
}

func (c extension) ExtensionName() string {
	return "ComplexityReporter"
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

	if complexityCalcs > c.threshold {
		c.complexityReporter.ReportComplexity(ctx, rc.OperationName, complexityCalcs)
	}

	return next(ctx)
}
