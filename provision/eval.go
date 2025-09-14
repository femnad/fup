package provision

import (
	"log/slog"
	"strings"

	"github.com/femnad/fup/entity"
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
			slog.Warn("Hint", "fact", fact, "message", hint.Message)
		}
	}

	return nil
}
