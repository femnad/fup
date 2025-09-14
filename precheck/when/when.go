package when

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"text/template"

	"github.com/femnad/fup/precheck"
)

type Whenable interface {
	RunWhen() string
}

func EvalStatement(statement string) (bool, error) {
	if statement == "" {
		return true, nil
	}

	tmpl := template.New("when").Funcs(precheck.FactFns)
	parsed, err := tmpl.Parse(fmt.Sprintf("{{%s}}", statement))
	if err != nil {
		return false, err
	}

	var out bytes.Buffer
	err = parsed.Execute(&out, context.TODO())
	if err != nil {
		return false, err
	}

	return out.String() == "true", nil
}

func ShouldRun(whenable Whenable) bool {
	statement := whenable.RunWhen()
	if statement == "" {
		// No statement defined, always should run.
		return true
	}

	shouldRun, err := EvalStatement(statement)
	if err != nil {
		slog.Warn("error evaluating check", "statement", statement, "error", err)
		return false
	}

	return shouldRun
}
