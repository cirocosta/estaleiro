package config

type AptKey struct {
	Name string `hcl:"name,label"`

	Uri string `hcl:"uri"`
}

type AptRepository struct {
	Name string `hcl:"name,label"`

	Uri    string `hcl:"uri"`
	Source string `hcl:"source,optional"`
}

// Package is a debian package to retrieve from the currently installed set of
// apt repos.
//
type Package struct {
	Name    string `hcl:"name,label"`
	Version string `hcl:"version,optional"`
}

func (p Package) String() string {
	res := p.Name

	if p.Version != "" {
		res = res + "=" + p.Version
	}

	return res
}

// Vcs represents information about a version control system that allows
// us to retrieve the source code from such file.
//
type VCS struct {
	Type string `hcl:"type,label"`

	Ref        string `hcl:"ref"`
	Repository string `hcl:"repository"`
}

// Step represents a build step that can be done using a regular
// Dockerfile (essentially, what `docker build` would build).
//
// Once referenced in an `image`, its DAG (until the target) gets merged
// into the tree that represents the build of the final image.
//
// Example:
//
// ```
//
// step "build" {
//   dockerfile = "./Dockerfile"
//   target     = "build"
// }
//
// ```
//
type Step struct {
	Name string ` hcl:"name,label"`

	Dockerfile string ` hcl:"dockerfile"`
	Target     string ` hcl:"target,optional"`
	Context    string ` hcl:"context,optional"`

	SourceFiles []struct {
		Location string `hcl:"location,label"`
		VCS      VCS    `hcl:"vcs,block"`
	} `hcl:"source_file,block"`
}

// File corresponds to a file that can be brought into the the final
// image from a diverse set of possible places (steps, images, local
// filesystem, etc).
//
// Example:
//
// ```
// # having a step named `build` ...
//
// file "/usr/local/bin/concourse" {
//   from_step "build" {
//     path = "/usr/local/concourse/bin/concourse"
//   }
// }
// ```
//
type File struct {
	Destination string ` hcl:"destination,label"`

	FromStep *struct {
		StepName string `hcl:"step_name,label"`
		Path     string `hcl:"path"`
	} `hcl:"from_step,block"`
}

// Config represents the high-level aggregation of all that's there to be built
// as a container image.
//
type Config struct {
	Image Image  `hcl:"image,block"`
	Steps []Step `hcl:"step,block"`
}

// BaseImage is the image that is going to be retrieved from a registry (or
// locally) to base the final layer of.
//
type BaseImage struct {
	Name      string `hcl:"name"`
	Reference string `hcl:"ref,optional"`
}

type Apt struct {
	Packages     []Package       `hcl:"package,block"`
	Repositories []AptRepository `hcl:"repository,block"`
	Keys         []AptKey        `hcl:"key,block"`
}

// Image is the final layer that is meant to be shipped as a container
// image.
//
type Image struct {
	Name string `hcl:"name,label" `

	BaseImage BaseImage `hcl:"base_image,block" `
	Apt       []Apt     `hcl:"apt,block"`

	Cmd        []string          `hcl:"cmd,optional"`
	Entrypoint []string          `hcl:"entrypoint"`
	Env        map[string]string `hcl:"env,optional"`
	Files      []File            `hcl:"file,block" `
	StopSignal string            `hcl:"stopsignal,optional"`
	Volumes    []string          `hcl:"volumes,optional"`
}
