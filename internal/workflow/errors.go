package workflow

import "errors"

var (
	ErrWorkflowNameEmpty = errors.New("workflow name is empty")
	ErrNoWorkflowSteps   = errors.New("workflow has no steps")
	ErrStepIDEmpty       = errors.New("step ID is empty")
)
