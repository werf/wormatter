package main

type Stage string

const (
	StagePreInstall  Stage = "pre-install"
	StageInstall     Stage = "install"
	StagePostInstall Stage = "post-install"
)

type ReleaseType string

const (
	ReleaseTypeMajor ReleaseType = "major"
	ReleaseTypeMinor ReleaseType = "minor"
	ReleaseTypePatch ReleaseType = "patch"
)

const UntypedZ = "z"
const UntypedA = "a"
const UntypedM = "m"

const untypedPrivateZ = "z"
const untypedPrivateA = "a"

func main() {}
