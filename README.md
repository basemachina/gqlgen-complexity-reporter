# gqlgen-complexity-reporter

This is a simple middleware of [gqlgen](https://gqlgen.com/) to measure the complexity of a GraphQL query.

## Installation

```bash
go get github.com/basemachina/gqlgen-complexity-reporter
```

## Usage

```go
package main

import (
    ...
    complexityreporter "github.com/basemachina/gqlgen-complexity-reporter"
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
    r := reporter{logger: logger}
    h.Use(complexityreporter.NewExtension(r, complexityreporter.WithThreshold(100))) // 100 is the maximum complexity allowed

    http.Handle("/", playground.Handler("GraphQL playground", "/query"))
    http.Handle("/query", auth.AuthMiddleware(srv))

    log.Fatal(http.ListenAndServe(":"+port, nil))
}
```
