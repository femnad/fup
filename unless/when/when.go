package when

import (
	"strings"

	"github.com/femnad/fup/internal"
	precheck "github.com/femnad/fup/unless"
)

type Whenable interface {
	RunWhen() string
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

	factFn, ok := precheck.Facts[fact]
	if !ok {
		internal.Log.Warningf("no fact evaluator for fact %s exists", fact)
		// Has a fact defined but we can't locate it, prefer not to run.
		return false
	}

	factResult, err := factFn()
	if err != nil {
		internal.Log.Errorf("error running evaluator for fact %s: %v", fact, err)
		// Has a fact defined, we can locate it but there's an error when evaluating it, prefer not to run.
		return false
	}

	if negate {
		return !factResult
	}
	return factResult
}
