package provision

import (
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/when"
)

func evalFacts(config entity.Config) error {
	for _, hint := range config.Hints {
		ok, err := when.EvalStatement(hint.Fact)
		if err != nil {
			return err
		}

		if !ok {
			internal.Log.Warningf("Hint for fact `%s`: %s", hint.Fact, hint.Message)
		}
	}

	return nil
}
