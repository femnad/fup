package when

import (
	"fmt"
	"strings"

	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck"
)

type Whenable interface {
	RunWhen() string
}

func FactOk(fact string) (bool, error) {
	factFn, ok := precheck.Facts[fact]
	if !ok {
		return false, fmt.Errorf("no fact evaluator for fact %s exists", fact)
	}

	factResult, err := factFn()
	if err != nil {
		return false, fmt.Errorf("error running evaluator for fact %s: %v", fact, err)
	}

	return factResult, nil
}

func ShouldRun(whenable Whenable) bool {
	negate := false
	fact := whenable.RunWhen()
	if fact == "" {
		// No fact defined, always should run.
		return true
	}

	tokens := strings.Split(fact, " ")
	if len(tokens) == 2 && tokens[0] == "not" {
		negate = true
		fact = tokens[1]
	}

	result, err := FactOk(fact)
	if err != nil {
		internal.Log.Warningf("error evaluating fact %s: %v", fact, err)
		return false
	}

	if negate {
		return !result
	}
	return result
}
