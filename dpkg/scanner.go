package dpkg

import (
	"bufio"
	"io"
	"strings"

	"github.com/pkg/errors"
)

type Scanner struct {
	scanner *bufio.Scanner
}

func NewScanner(reader io.Reader) Scanner {
	return Scanner{
		scanner: bufio.NewScanner(reader),
	}
}

func (p *Scanner) ScanAll() (pkgs []DebControl, err error) {
	var (
		pkg  DebControl
		done bool
	)

	for {
		pkg, done, err = p.Scan()
		if err != nil {
			err = errors.Wrapf(err, "failed scanning content from dpkg status")
			return
		}

		if done {
			return
		}

		pkgs = append(pkgs, pkg)

	}

	return
}

func (p *Scanner) Scan() (pkg DebControl, done bool, err error) {
	for {
		done = !p.scanner.Scan()
		if done {
			err = p.scanner.Err()
			return
		}

		line := p.scanner.Text()

		if len(line) == 0 {
			return
		}

		if strings.HasPrefix(line, " ") {
			continue
		}

		splitted := strings.SplitN(line, ":", 2)
		if len(splitted) != 2 {
			err = errors.Errorf("failed parsing `k:v` in line `%s`", line)
			return
		}

		key, value := splitted[0], strings.TrimSpace(splitted[1])
		switch key {
		case "Package":
			pkg.Name = value
		case "Source":
			pkg.SourcePackage = value
		case "Version":
			pkg.Version = value
		case "Architecture":
			pkg.Architecture = value
		case "Maintainer":
			pkg.Maintainer = value
		case "Description":
			pkg.Description = value
		case "Depends":
			pkg.Depends = value
		case "Pre-Depends":
			pkg.PreDepends = value
		case "Recommends":
			pkg.Recommends = value
		case "Suggests":
			pkg.Suggests = value
		case "Breaks":
			pkg.Breaks = value
		case "Conflicts":
			pkg.Conflicts = value
		case "Provides":
			pkg.Provides = value
		case "Replaces":
			pkg.Replaces = value
		case "Enhances":
			pkg.Enhances = value

		}
	}

	return
}
