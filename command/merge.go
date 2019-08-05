package command

type mergeCommand struct {
	Directory string `long:"directory" required:"true" description:"where bill of material files live"`
	Output    string `long:"output"    required:"true" description:"where to write the bill of materials to ('-' for stdout)"`
}

func (c *mergeCommand) Execute(args []string) (err error) {
	// list all files in the directory
	// interpret them
	// perform the merging
	// write the final bom to a `bom.yml`

	return
}
