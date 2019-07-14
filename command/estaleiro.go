package command

var Estaleiro struct {
	Build    buildCommand    `command:"build" description:"runs a build"`
	Frontend frontendCommand `command:"serves as a custom Docker-compatible frontend"`
}
