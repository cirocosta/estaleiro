package command

var Estaleiro struct {
	LLB      llbCommand      `command:"llb" description:"generates the llb for a build"`
	Build    buildCommand    `command:"build" description:"performs a build against buildkitd"`
	Frontend frontendCommand `command:"frontend" description:"serves as a custom Docker-compatible"`

	Collect collectCommand `command:"collect" hidden:"true" description:"retrieves dpkg packages information"`
	Apt     aptCommand     `command:"apt" hidden:"true"`
}
