# gqlgen-complexity-metrics

This is a simple middleware of [gqlgen](https://gqlgen.com/) to measure the complexity of a GraphQL query.

## Installation

```bash
go get github.com/basemachina/gqlgen-complexity-metrics
```

## Usage

```go
package main

import (
    ...
    complexitymetrics "github.com/basemachina/gqlgen-complexity-metrics"
)

type reporter struct {
    logger *zap.Logger
}

func (r *reporter) ReportComplexity(ctx context.Context, operationName string, complexity int) {
    r.logger.Info("[graphql query complexity]", zap.Int("complexity", complexity))
}

func main() {
    srv := handler.NewDefaultServer(internal.NewExecutableSchema(internal.Config{
        ...
    }))
    logger, _ := zap.NewProduction()
    h.Use(complexitymetrics.NewComplexityReporterExtension(100, reporter{logger: logger})) // 100 is the maximum complexity allowed

    http.Handle("/", playground.Handler("GraphQL playground", "/query"))
    http.Handle("/query", auth.AuthMiddleware(srv))

    log.Fatal(http.ListenAndServe(":"+port, nil))
}
```
