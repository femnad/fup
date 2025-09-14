package provision

import (
	"errors"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/precheck/when"
)

const (
	hintBase = "Hint for running task"
)

func runTask(task entity.Task, cfg entity.Config) error {
	name := task.Name()

	if !when.ShouldRun(task) {
		if task.Hint != "" && !unless.ShouldSkip(task, cfg.Settings) {
			internal.Logger.Warn().Str("name", name).Str("task", task.Hint).Msg(hintBase)
		}
		internal.Logger.Trace().Str("name", name).Str("when", task.When).Msg("Skipping task")
		return nil
	}

	if unless.ShouldSkip(task, cfg.Settings) {
		internal.Logger.Trace().Str("name", name).Msg("Skipping task")
		return nil
	}

	internal.Logger.Info().Str("name", name).Msg("Running task")
	return task.Run(cfg)
}

func runTasks(cfg entity.Config, tasks []entity.Task) error {
	var taskErrs []error

	for _, task := range tasks {
		err := runTask(task, cfg)
		taskErrs = append(taskErrs, err)
	}

	return errors.Join(taskErrs...)
}
