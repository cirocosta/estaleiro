package config

// Config represents the high-level aggregation of all that's there to be built
// as a container image, including the sources that bring files to it, and the
// final definition itself.
//
type Config struct {
	Image    Image     `hcl:"image,block"`
	Steps    []Step    `hcl:"step,block"`
	Tarballs []Tarball `hcl:"tarball,block"`
}

// Image is the final layer that is meant to be shipped as a container
// image, specifying the desired final intent.
//
type Image struct {
	Name string `hcl:"name,label" `

	BaseImage BaseImage `hcl:"base_image,block" `

	// State providers
	//

	Apt   []Apt  `hcl:"apt,block"`
	Files []File `hcl:"file,block" `

	// Final image configuration
	//

	Cmd        []string          `hcl:"cmd,optional"`
	Entrypoint []string          `hcl:"entrypoint,optional"`
	Env        map[string]string `hcl:"env,optional"`
	StopSignal string            `hcl:"stopsignal,optional"`
	Volumes    []string          `hcl:"volumes,optional"`
}

// Step represents a build step that can be done using a regular
// Dockerfile (essentially, what `docker build` would build).
//
// Once referenced in an `image`, its DAG (until the target) gets merged
// into the tree that represents the build of the final image.
//
type Step struct {
	Name string `hcl:"name,label"`

	Dockerfile string `hcl:"dockerfile"`
	Target     string `hcl:"target,optional"`
	Context    string `hcl:"context,optional"`

	SourceFiles []SourceFile `hcl:"source_file,block"`
}

// Tarball represents a compressed tarball that contains files that can be
// consumed by the image.
//
// ps.: a file must be declarated before consumption by the image.
//
type Tarball struct {
	// Name corresponds to the path in the context where the tarball can be
	// found.
	//
	Name string `hcl:"name,label"`

	SourceFiles []SourceFile `hcl:"source_file,block"`
}

type SourceFile struct {
	Location string `hcl:"location,label"`
	VCS      VCS    `hcl:"vcs,block"`
}

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

type TarballFile struct {
	Name  string   `hcl:"name"`
	Paths []string `hcl:"paths"`
	VCS   VCS      `hcl:"vcs,block"`
}

type FileFromStep struct {
	StepName string `hcl:"step_name,label"`
	Path     string `hcl:"path"`
}

type FileFromTarball struct {
	TarballName string `hcl:"tarball_name,label"`
	Path        string `hcl:"path"`
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

	FromStep    *FileFromStep    `hcl:"from_step,block"`
	FromTarball *FileFromTarball `hcl:"from_tarball,block"`
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
