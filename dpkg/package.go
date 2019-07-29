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
	Description  string `yaml:"-"`
	Maintainer   string `yaml:"-"`

	// interrelationship fields that describe the package's relationship
	// with other packages
	//
	// ref: https://www.debian.org/doc/debian-policy/ch-controlfields.html
	//

	Depends    string `yaml:"-"`
	PreDepends string `yaml:"-"`
	Recommends string `yaml:"-"`
	Suggests   string `yaml:"-"`
	Breaks     string `yaml:"-"`
	Conflicts  string `yaml:"-"`
	Provides   string `yaml:"-"`
	Replaces   string `yaml:"-"`
	Enhances   string `yaml:"-"`

	// fields that are mandatory for making the package serveable as part
	// of a debian depository.
	//
	// ref: https://wiki.debian.org/DebianRepository/Format#A.22Release.22_files
	//

	Filename string `yaml:"-"`
	Size     string `yaml:"-"`
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

	if d.Depends != "" {
		str += "Depends: " + d.Depends + "\n"
	}

	if d.PreDepends != "" {
		str += "Pre-Depends: " + d.PreDepends + "\n"
	}

	if d.Recommends != "" {
		str += "Recommends: " + d.Recommends + "\n"
	}

	if d.Suggests != "" {
		str += "Suggests: " + d.Suggests + "\n"
	}

	if d.Breaks != "" {
		str += "Breaks: " + d.Breaks + "\n"
	}

	if d.Conflicts != "" {
		str += "Conflicts: " + d.Conflicts + "\n"
	}

	if d.Provides != "" {
		str += "Provides: " + d.Provides + "\n"
	}

	if d.Replaces != "" {
		str += "Replaces: " + d.Replaces + "\n"
	}

	if d.Enhances != "" {
		str += "Enhances: " + d.Enhances + "\n"
	}

	str += "\n"

	return
}
