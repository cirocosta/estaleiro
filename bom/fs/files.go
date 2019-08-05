package fs

type UnarchivedTarball struct {
	// Path to the original tarball
	//
	Path string `yaml:"path"`

	// Digest of the original tarball
	//
	Digest string `yaml:"digest"`

	// UnarchivedLocation corresponds to where the contents of the tarball
	// where unarchived to.
	//
	UnarchivedLocation string `yaml:"unarchived_location"`
}

type ExtractedFile struct {
	Name              string            `yaml:"name"`
	Path              string            `yaml:"path"`
	Digest            string            `yaml:"digest"`
	UnarchivedTarball UnarchivedTarball `yaml:"from_tarball"`
}

type FilesV1 struct {
	Kind string          `yaml:"kind"`
	Data []ExtractedFile `yaml:"data"`
}

func NewFilesV1(extractedFiles []ExtractedFile) FilesV1 {
	return FilesV1{
		Kind: "files/v1",
		Data: extractedFiles,
	}
}
