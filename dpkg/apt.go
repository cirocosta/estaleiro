package dpkg

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type AptDebLocation struct {
	URI    string `yaml:"uri"`
	Name   string `yaml:"name"`
	Size   string `yaml:"size"`
	Digest string `yaml:"digest"`
}

type Package struct {
	DebControl
	Source AptDebLocation `yaml:"source,omitempty"`
}

func InstallPackages(ctx context.Context, packages []string) (pkgs []Package, err error) {
	locations, err := aptInstallPackagesUris(ctx, packages)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to install packages %v", packages)
		return
	}

	dir, err := ioutil.TempDir("", "estaleiro-deb-packages")
	if err != nil {
		err = errors.Wrapf(err,
			"failed creating temp directory for debian packages")
		return
	}

	err = downloadDebianPackages(ctx, dir, locations)
	if err != nil {
		err = errors.Wrapf(err,
			"failed downloading packages %v", packages)
		return
	}

	err = installDebianPackages(ctx, dir)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to install downloaded debian packages")
		return
	}

	return
}

func getDebianPackageInfo(ctx context.Context, filename string) (pkg DebControl, err error) {
	var (
		cmd = exec.CommandContext(ctx,
			"dpkg-deb", "--info", filename, "control")
		out bytes.Buffer
	)

	cmd.Stderr = &out
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to retrieve debian package information for %s - %s",
			filename, string(out.Bytes()))
		return
	}

	var (
		scanner = NewScanner(&out)
		done    bool
	)
	pkg, done, err = scanner.Scan()
	if err != nil {
		err = errors.Wrapf(err,
			"failed scaning pkg info from `dpkg-deb --info` on %s", filename)
		return
	}

	if done {
		err = errors.Errorf("no package info retrieved for %s", filename)
		return
	}

	return
}

func installDebianPackages(ctx context.Context, dir string) (err error) {
	var cmd = exec.CommandContext(ctx,
		"dpkg", "--install", "--recursive", dir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to install debian packages from directory %s - %s",
			dir, string(out))
		return
	}

	return
}

func aptInstallPackagesUris(ctx context.Context, packages []string) ([]AptDebLocation, error) {
	return aptUris(ctx, "install", packages)
}

func aptSourcePackagesUris(ctx context.Context, packages []string) ([]AptDebLocation, error) {
	return aptUris(ctx, "source", packages)
}

func aptUris(ctx context.Context, command string, packages []string) (uris []AptDebLocation, err error) {
	var (
		stdout, stderr bytes.Buffer

		cmd = exec.CommandContext(ctx, "apt-get", append([]string{
			"--print-uris",
			command},
			packages...)...)
	)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to retrieve uris for packages %v - %s",
			packages, string(stderr.Bytes()))
		return
	}

	uris, err = ScanAptDebLocations(&stdout)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to scan packages uris")
		return
	}

	return
}

func downloadDebianPackages(ctx context.Context, dir string, locations []AptDebLocation) (err error) {
	var (
		eg *errgroup.Group
	)

	eg, ctx = errgroup.WithContext(ctx)

	for _, location := range locations {
		location := location

		eg.Go(func() error {
			return downloadDebianPackage(ctx, dir, location)
		})
	}

	err = eg.Wait()
	if err != nil {
		err = errors.Wrapf(err,
			"failed during debian packages retrieval")
		return
	}

	return
}

func downloadDebianPackage(ctx context.Context, dir string, location AptDebLocation) (err error) {
	out, err := os.Create(path.Join(dir, location.Name))
	if err != nil {
		err = errors.Wrapf(err,
			"failed creating destination for debian package %s", location.Name)
		return err
	}
	defer out.Close()

	req, err := http.NewRequest("GET", location.URI, nil)
	if err != nil {
		err = errors.Wrapf(err,
			"failed creating request to retrieve debian package '%s' at '%s'",
			location.Name, location.URI)
		return
	}

	req = req.WithContext(ctx)

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to submit request to retrieve deb package at %s", location.URI)
		return
	}

	defer res.Body.Close()

	_, err = io.Copy(out, res.Body)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to write body to file '%s' from request to '%s'",
			location.Name, location.URI)
		return
	}

	return
}

func ScanAptDebLocations(reader io.Reader) (locations []AptDebLocation, err error) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, `'http`) {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 4 {
			err = errors.Errorf("malformed line `%s`", line)
			return
		}

		locations = append(locations, AptDebLocation{
			URI:    strings.Trim(fields[0], "'"),
			Name:   fields[1],
			Size:   fields[2],
			Digest: fields[3],
		})
	}

	return
}
