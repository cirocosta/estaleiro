package dpkg

// DebControl is a debian package whose installation can be tracked by
// `/var/lib/dpkg/status`.
//
// For more information about each field, check:
// https://www.debian.org/doc/debian-policy/ch-controlfields#list-of-fields
//
//
type DebControl struct {
	// Name corresponds to the `Package` field in the status file from
	// `dpkg`, representing the name of the binary package.
	//
	Name string `yaml:"name"`

	// Version is the version of the package in the debian policy format
	// (`Version` field)
	//
	Version string `yaml:"version"`

	// SourcePackage is a field that gets set for binary packages
	// indicating what's the name of the source package that brings this
	// one. (`Source` field)
	//
	SourcePackage string `yaml:"source_package"`

	Architecture string `yaml:"architecture"`

	Maintainer  string
	Description string
}

// ControlString generates a multiline-string that corresponds to the format
// that debian control files adhere to.
//
// Mandatory in either binary or source packages:
//
// - Architecture
// - Description
// - Maintainer
// - Package
// - Version
//
func (d DebControl) ControlString() (str string) {
	str += "Package: " + d.Name + "\n"

	if d.SourcePackage != "" {
		str += "Source: " + d.SourcePackage + "\n"
	}

	str += "Architecture: " + d.Architecture + "\n"
	str += "Description: " + d.Description + "\n"
	str += "Maintainer: " + d.Maintainer + "\n"
	str += "Version: " + d.Version + "\n"

	str += "\n"

	return
}
