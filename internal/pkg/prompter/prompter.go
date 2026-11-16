package prompter

import "github.com/charmbracelet/huh"

// Option is a selectable choice: Label is shown to the user, Value is returned.
type Option struct {
	Label string
	Value string
}

type Prompter interface {
	Text(title string) (string, error)
	Secret(title string) (string, error)
	Select(title string, options []Option) (string, error)
}

type huhPrompter struct{}

func NewHuhPrompter() Prompter {
	return &huhPrompter{}
}

func (p *huhPrompter) Text(title string) (string, error) {
	var v string
	err := huh.NewInput().Title(title).Prompt(">").Value(&v).Run()

	return v, err
}

func (p *huhPrompter) Secret(title string) (string, error) {
	var v string
	err := huh.NewInput().Title(title).Prompt(">").EchoMode(huh.EchoModePassword).Value(&v).Run()

	return v, err
}

func (p *huhPrompter) Select(title string, options []Option) (string, error) {
	var v string

	opts := make([]huh.Option[string], 0, len(options))
	for _, o := range options {
		opts = append(opts, huh.NewOption(o.Label, o.Value))
	}

	err := huh.NewSelect[string]().Title(title).Options(opts...).Value(&v).Run()

	return v, err
}
