package command

var Estaleiro struct {
	Build    buildCommand    `command:"build" description:"runs a build"`
	Frontend frontendCommand `command:"frontend" description:"serves as a custom Docker-compatible"`
}
