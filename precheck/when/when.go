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

func evalFact(fact string, negate bool) (bool, error) {
	result, err := FactOk(fact)
	if err != nil {
		return false, fmt.Errorf("error evaluating fact %s: %v", fact, err)
	}

	if negate {
		result = !result
	}
	return result, nil
}

func ShouldRun(whenable Whenable) bool {
	negate := false
	fact := whenable.RunWhen()
	if fact == "" {
		// No fact defined, always should run.
		return true
	}

	if strings.HasPrefix(fact, "!") {
		fact = fact[1:]
		negate = true
	}

	for _, subFact := range strings.Split(fact, " ") {
		result, err := evalFact(subFact, negate)
		if err != nil {
			internal.Log.Warningf("error evaluating fact %s: %v", subFact, err)
			return false
		}

		return result
	}

	// No facts found
	return false
}
