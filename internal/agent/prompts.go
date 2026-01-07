package agent

import (
	"context"
	_ "embed"

	"github.com/trankhanh040147/revcli/internal/agent/prompt"
	"github.com/trankhanh040147/revcli/internal/config"
)

//go:embed templates/reviewer.md.tpl
var reviewerPromptTmpl []byte

//go:embed templates/coder.md.tpl
var coderPromptTmpl []byte

// go:embed templates/planner.md.tpl
var plannerPromptTmpl []byte

//go:embed templates/task.md.tpl
var taskPromptTmpl []byte

//go:embed templates/initialize.md.tpl
var initializePromptTmpl []byte

// TODO(PlanC): Add structured review output format (severity levels, categories)
func reviewerPrompt(opts ...prompt.Option) (*prompt.Prompt, error) {
	systemPrompt, err := prompt.NewPrompt("reviewer", string(reviewerPromptTmpl), opts...)
	if err != nil {
		return nil, err
	}
	return systemPrompt, nil
}

// TODO(PlanC): Rename to builderPrompt for build mode
func coderPrompt(opts ...prompt.Option) (*prompt.Prompt, error) {
	systemPrompt, err := prompt.NewPrompt("coder", string(coderPromptTmpl), opts...)
	if err != nil {
		return nil, err
	}
	return systemPrompt, nil
}

// TODO(PlanC): Add comparison mode (diff-based, commit-based reviews)
func taskPrompt(opts ...prompt.Option) (*prompt.Prompt, error) {
	systemPrompt, err := prompt.NewPrompt("task", string(taskPromptTmpl), opts...)
	if err != nil {
		return nil, err
	}
	return systemPrompt, nil
}

func plannerPrompt(opts ...prompt.Option) (*prompt.Prompt, error) {
	systemPrompt, err := prompt.NewPrompt("planner", string(plannerPromptTmpl), opts...)
	if err != nil {
		return nil, err
	}
	return systemPrompt, nil
}

func InitializePrompt(cfg config.Config) (string, error) {
	systemPrompt, err := prompt.NewPrompt("initialize", string(initializePromptTmpl))
	if err != nil {
		return "", err
	}
	return systemPrompt.Build(context.Background(), "", "", cfg)
}

// todo: explain agent package
