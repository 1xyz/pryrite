package components

import (
	"github.com/manifoldco/promptui"
	"strings"
)

func ShowYNQuestionPrompt(question string) bool {
	prompt := promptui.Prompt{
		Label:     question,
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		return false
	}
	return strings.ToLower(result) == "y"
}
