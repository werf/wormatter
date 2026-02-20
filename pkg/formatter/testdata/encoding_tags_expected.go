package main

type Operation struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Action string `json:"action"`
}

type Plan struct {
	Version int          `yaml:"version"`
	Name    string       `yaml:"name"`
	Steps   []*Operation `yaml:"steps"`
}

type NoTags struct {
	Alpha string
	Zulu  string

	bravo int
}

type WithValidate struct {
	Alpha string `validate:"required"`
	Zulu  string `validate:"required"`

	bravo int
}

func main() {}
