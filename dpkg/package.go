package dpkg

// Package is a debian package whose installation can be tracked by
// `/var/lib/dpkg/status`.
//
// For more information about each field, check:
// https://www.debian.org/doc/debian-policy/ch-controlfields#list-of-fields
//
//
type Package struct {
	// Name corresponds to the `Package` field in the status file from
	// `dpkg`, representing the name of the binary package.
	//
	Name string `yaml:"name"`

	// Version is the version of the package in the debian policy format.
	//
	Version string `yaml:"version"`
}

func (p *Package) IsFilled() bool {
	return p.Name != "" && p.Version != ""
}
