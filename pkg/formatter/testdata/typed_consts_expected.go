package main

const (
	ReleaseTypeMajor ReleaseType = "major"
	ReleaseTypeMinor ReleaseType = "minor"
	ReleaseTypePatch ReleaseType = "patch"

	StagePreInstall  Stage = "pre-install"
	StageInstall     Stage = "install"
	StagePostInstall Stage = "post-install"

	UntypedA = "a"
	UntypedM = "m"
	UntypedZ = "z"

	untypedPrivateA = "a"
	untypedPrivateZ = "z"
)

type Stage string

type ReleaseType string

func main() {}
