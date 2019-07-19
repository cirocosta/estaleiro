package config

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
//
type Step struct {
	Name       string `hcl:"name,label"`
	Dockerfile string `hcl:"dockerfile"`
	Target     string `hcl:"target,optional"`
	Context    string `hcl:"context,optional"`
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
	Destination string `hcl:"destination,label"`

	FromStep *struct {
		StepName string `hcl:"step_name,label"`
		Path     string `hcl:"path"`
	} `hcl:"from_step,block"`
}

type Config struct {
	Image Image  `hcl:"image,block"`
	Steps []Step `hcl:"step,block"`
}

// Image is the final layer that is meant to be shipped as a container
// image.
//
type Image struct {
	Name      string `hcl:"name,label"`
	BaseImage struct {
		Name      string `hcl:"name"`
		Reference string `hcl:"ref,optional"`
	} `hcl:"base_image,block"`
	Files []File `hcl:"file,block"`
}
