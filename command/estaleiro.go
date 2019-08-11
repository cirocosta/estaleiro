// Package command defines the commands that can be leverage from `estaleiro`.
//
package command

import (
	"os"

	"code.cloudfoundry.org/lager"
)

var (
	logger = lager.NewLogger("estaleiro")
)

func init() {
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))
}

var Estaleiro struct {
	LLB      llbCommand      `command:"llb" description:"generates the llb for a build"`
	Build    buildCommand    `command:"build" description:"performs a build against buildkitd"`
	Frontend frontendCommand `command:"frontend" description:"serves as a custom Docker-compatible"`

	AptInstall      aptInstallCommand      `command:"apt-install"      hidden:"true"`
	AptRepositories aptRepositoriesCommand `command:"apt-repositories" hidden:"true"`
	Base            baseCommand            `command:"base"             hidden:"true"`
	Collect         collectCommand         `command:"collect"          hidden:"true"`
	Extract         extractCommand         `command:"extract"          hidden:"true"`
	Merge           mergeCommand           `command:"merge"            hidden:"true"`
}
