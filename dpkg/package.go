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

	// Version is the version of the package in the debian policy format.
	//
	Version string `yaml:"version"`

	// SourcePackage is a field that gets set for binary packages indicating
	// what's the name of the source package that brings this one.
	//
	SourcePackage string `yaml:"source_package"`
}
