package osrelease

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type OsRelease struct {
	OS       string `yaml:"os"`
	Version  string `yaml:"version"`
	Codename string `yaml:"codename"`
}

func GatherOsRelease() (info OsRelease, err error) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		err = errors.Wrapf(err,
			"failed to open `/etc/os-release`")
		return
	}
	defer f.Close()

	info = ScanInfo(f)

	return
}

func ScanInfo(reader io.Reader) (info OsRelease) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, "=")

		k, v := fields[0], fields[1]
		v = strings.Trim(v, `"`)

		switch k {
		case "ID":
			info.OS = v
		case "VERSION_ID":
			info.Version = v
		case "VERSION_CODENAME":
			info.Codename = v
		}
	}

	return
}
