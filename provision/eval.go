package provision

import (
	"strings"

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
			fact := strings.Replace(hint.Fact, "\"", "`", -1)
			internal.Logger.Warn().Str("hint", fact).Msg(hint.Message)
		}
	}

	return nil
}
