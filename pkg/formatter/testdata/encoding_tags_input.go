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
	Zulu  string
	Alpha string
	bravo int
}

type WithValidate struct {
	Zulu  string `validate:"required"`
	Alpha string `validate:"required"`
	bravo int
}

func main() {}
