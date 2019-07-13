package config

// Step represents a build step that can be done using a regular Dockerfile
// (essentially, what `docker build` would build).
//
// Once referenced in an `image`, its DAG (until the target) gets merged into
// the tree that represents the build of the final image.
//
type Step struct {
	Name       string `hcl:"name,label"`
	Dockerfile string `hcl:"dockerfile"`
	Target     string `hcl:"target,optional"`
	Context    string `hcl:"context,optional"`
}

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

type Image struct {
	Name      string `hcl:"name,label"`
	BaseImage struct {
		Name      string `hcl:"name"`
		Reference string `hcl:"ref,optional"`
	} `hcl:"base_image,block"`
	Files []File `hcl:"file,block"`
}
