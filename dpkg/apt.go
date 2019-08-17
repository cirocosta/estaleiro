package dpkg

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"code.cloudfoundry.org/lager"
	"github.com/cirocosta/estaleiro/osrelease"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var (
	logger = lager.NewLogger("estaleiro")
)

func init() {
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))
}

func CreateDebianRepositoryIndex(dir string, pkgs []Package) (err error) {
	var buffer bytes.Buffer

	for _, pkg := range pkgs {
		_, err = buffer.WriteString(pkg.DebControl.ControlString())
		if err != nil {
			err = errors.Wrapf(err,
				"failed writing package control string to memory buffer")
			return
		}
	}

	filename := path.Join(dir, "Packages")

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		err = errors.Wrapf(err,
			"failed creating Packages file in %s",
			filename)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, &buffer)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to write control string buffer to file %s", filename)
		return
	}

	return
}

func ComputeSHA256(filename string) (sum string, err error) {
	f, err := os.Open(filename)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to open file %s when compute sha256", filename)
		return
	}
	defer f.Close()

	h := sha256.New()

	_, err = io.Copy(h, f)
	if err != nil {
		err = errors.Wrapf(err,
			"failed copying contents from file %s to sha256 hasher", filename)
		return
	}

	sum = fmt.Sprintf("%x", h.Sum(nil))
	return

}

func shouldSkipSource(name string, skipSources []string) bool {
	for _, s := range skipSources {
		if name == s {
			return true
		}
	}

	return false
}

func CreatePackages(
	ctx context.Context, dir string, locations []AptDebLocation, skipSource []string,
) (
	pkgs []Package, err error,
) {
	var eg *errgroup.Group

	pkgs = make([]Package, len(locations))
	eg, ctx = errgroup.WithContext(ctx)

	for idx, location := range locations {
		idx, location := idx, location

		eg.Go(func() (err error) {
			control, err := getDebianPackageInfo(ctx, path.Join(dir, location.Name))
			if err != nil {
				return
			}

			control.Filename = location.Name
			control.Size = location.Size
			control.MD5sum = location.MD5sum
			control.SHA256, err = ComputeSHA256(path.Join(dir, location.Name))
			if err != nil {
				return
			}

			// how to conditionally check for this?

			var (
				sourceLocations []AptDebLocation
				ref             = control.Name + "=" + control.Version
			)

			if shouldSkipSource(ref, skipSource) {
				sourceLocations = []AptDebLocation{}
			} else {
				sourceLocations, err = aptSourcePackageUris(ctx, ref)
				if err != nil {
					err = errors.Wrapf(err,
						"failed to retrieve source for package %s",
						ref)
					return
				}
			}

			pkgs[idx] = Package{
				DebControl: control,
				Location:   location,
				Source:     sourceLocations,
			}

			return
		})
	}

	err = eg.Wait()
	if err != nil {
		err = errors.Wrapf(err,
			"failed retrieving debian packages info")
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

	logger.Info("get-deb-pkg-info", lager.Data{"filename": filename})

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
	)

	pkg, _, err = scanner.Scan()
	if err != nil {
		err = errors.Wrapf(err,
			"failed scaning pkg info from `dpkg-deb --info` on %s", filename)
		return
	}

	return
}

// ForceLocalSource ensures that we're only able to retrieve debian packages
// from the directory where we downloaded our stuff to so that no other
// repositories can provide those (which wouldn't be appropriately tracked).
//
// as this involves performing changes in the filesystem, this may lead to a
// system without proper configuration for `apt`.
//
// ps.: backup files/dirs are kept with `-backup` suffixed names.
//
func ForceLocalSourcesList(ctx context.Context, repositoryDir string) (err error) {
	logger.Info("force-local-sources", lager.Data{"repository-dir": repositoryDir})

	err = ioutil.WriteFile("/etc/apt/sources.list",
		[]byte("deb [trusted=yes allow-weak-repositories=yes] file:"+repositoryDir+" ./"),
		0644,
	)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to write local apt source repository to sources.list")
		return
	}

	err = RemoveAptLists()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to remove apt repository listing")
		return
	}

	err = UpdateApt(ctx)
	if err != nil {
		err = errors.Wrapf(err,
			"couldn't update apt repositories")
		return
	}

	return
}

func RemoveAptLists() error {
	return os.RemoveAll("/var/lib/apt/lists")
}

func UpdateApt(ctx context.Context) (err error) {
	logger.Info("update-apt")

	var cmd = exec.CommandContext(ctx, "apt", "update")

	out, err := cmd.CombinedOutput()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to perform `apt update` %s", string(out))
		return
	}

	return
}

func InstallDebianPackages(ctx context.Context, packages []string) (err error) {
	args := append([]string{
		"install", "--yes", "--no-install-recommends", "--no-install-suggests",
	}, packages...)

	out, err := exec.CommandContext(ctx, "apt", args...).CombinedOutput()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to install packages %v - %s", args[4:], string(out))
		return
	}

	return
}

func GatherDebLocations(ctx context.Context, packages []string) ([]AptDebLocation, error) {
	return aptUris(ctx, "install", packages)
}

func aptSourcePackageUris(ctx context.Context, packageName string) (locations []AptDebLocation, err error) {
	return aptUris(ctx, "source", []string{packageName})
}

func aptUris(ctx context.Context, command string, packages []string) (uris []AptDebLocation, err error) {
	var (
		stdout, stderr bytes.Buffer

		cmd = exec.CommandContext(ctx, "apt-get", append([]string{
			"--print-uris",
			"--no-install-recommends",
			"--no-install-suggests",
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

// There are four official repositories that one
// can retrieve directly from ubuntu:
//
//
//                | FREE     | NON-FREE
//    ------------+----------+-----------
//      SUPPORTED | main     | restricted
//    -----------------------------------
//    UNSUPPORTED | universe | multiverse
//
func UbuntuSupportedRepositories() (res []string, err error) {
	info, err := osrelease.GatherOsRelease()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to detect codename to be used for initial sources.list")
		return
	}

	var repos = []string{
		"deb http://archive.ubuntu.com/ubuntu/ %s main restricted",
		"deb-src http://archive.ubuntu.com/ubuntu/ %s main restricted",

		"deb http://archive.ubuntu.com/ubuntu/ %s-updates main restricted",
		"deb-src http://archive.ubuntu.com/ubuntu/ %s-updates main restricted",

		"deb http://archive.ubuntu.com/ubuntu/ %s-security main restricted",
		"deb-src http://archive.ubuntu.com/ubuntu/ %s-security main restricted",
	}

	res = make([]string, len(repos))

	for idx, repo := range repos {
		res[idx] = fmt.Sprintf(repo, info.Codename)
	}

	return
}

func WriteSourcesList(repositories []string) (err error) {
	logger.Info("write-initial-list", lager.Data{
		"repositories": repositories,
	})

	f, err := os.OpenFile("/etc/apt/sources.list", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to open `/etc/apt/sources.list` file to write initial apt sources list")
		return
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, strings.Join(repositories, "\n"))
	if err != nil {
		err = errors.Wrapf(err,
			"failed writing initial sources.list to `/etc/apt/sources.list`")
		return
	}

	return
}

func DownloadDebianPackages(ctx context.Context, dir string, locations []AptDebLocation) (err error) {
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

func Download(ctx context.Context, uri, dest string) (err error) {
	out, err := os.Create(dest)
	if err != nil {
		err = errors.Wrapf(err,
			"failed destination file %s", dest)
		return err
	}
	defer out.Close()

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		err = errors.Wrapf(err,
			"failed creating request retrieve file %s", uri)
		return
	}

	req = req.WithContext(ctx)

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to submit request to retrieve file %s", uri)
		return
	}

	defer res.Body.Close()

	_, err = io.Copy(out, res.Body)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to write body to file '%s' from '%s'",
			dest, uri)
		return
	}

	return
}

func downloadDebianPackage(ctx context.Context, dir string, location AptDebLocation) (err error) {
	logger.Info("download-debian-package", lager.Data{"name": location.Name})

	err = Download(
		ctx,
		location.URI,
		path.Join(dir, location.Name),
	)
	if err != nil {
		err = errors.Wrapf(err,
			"failed downloading debian package %s", location.Name)
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

		digestFields := strings.Split(fields[3], ":")
		if len(digestFields) != 2 {
			err = errors.Errorf("malformed digest field: `%s`", fields[3])
			return
		}

		locations = append(locations, AptDebLocation{
			URI:    strings.Trim(fields[0], "'"),
			Name:   fields[1],
			Size:   fields[2],
			MD5sum: strings.TrimSpace(digestFields[1]),
		})
	}

	return
}
