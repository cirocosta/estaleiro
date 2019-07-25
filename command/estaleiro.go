package command

var Estaleiro struct {
	LLB      llbCommand      `command:"llb" description:"generates the llb for a build"`
	Frontend frontendCommand `command:"frontend" description:"serves as a custom Docker-compatible"`
}
