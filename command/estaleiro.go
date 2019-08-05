package command

var Estaleiro struct {
	LLB      llbCommand      `command:"llb" description:"generates the llb for a build"`
	Build    buildCommand    `command:"build" description:"performs a build against buildkitd"`
	Frontend frontendCommand `command:"frontend" description:"serves as a custom Docker-compatible"`

	Apt     aptCommand     `command:"apt"     hidden:"true"`
	Base    baseCommand    `command:"base"    hidden:"true"`
	Collect collectCommand `command:"collect" hidden:"true"`
	Extract extractCommand `command:"extract" hidden:"true"`
	Merge   mergeCommand   `command:"merge"   hidden:"true"`
}
