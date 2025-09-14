package provision

import (
	"errors"
	"log/slog"

	"github.com/femnad/fup/entity"
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
			slog.Warn(hintBase, "name", name, "task", task.Hint)
		}
		slog.Debug("Skipping running task as condition evaluated to false", "name", name, "when", task.When)
		return nil
	}

	if unless.ShouldSkip(task, cfg.Settings) {
		slog.Debug("Skipping running task as condition evaluated to true", "name", name, "unless",
			task.Unless)
		return nil
	}

	slog.Info("Running task", "name", name)
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
