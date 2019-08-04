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
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type AptDebLocation struct {
	URI    string `yaml:"uri"`
	Name   string `yaml:"name"`
	Size   string `yaml:"size"`
	MD5sum string `yaml:"md5sum"`
}

type Package struct {
	DebControl `yaml:",inline"`
	Location   AptDebLocation   `yaml:"location,omitempty"`
	Source     []AptDebLocation `yaml:"source,omitempty"`
}

var (
	logger = lager.NewLogger("estaleiro")
)

func init() {
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))
}

func InstallPackages(ctx context.Context, packages []string) (pkgs []Package, err error) {
	logger.Info("install-packages", lager.Data{"packages": packages})

	err = writeInitialList()
	if err != nil {
		err = errors.Wrapf(err,
			"failed setting up initial sources.list")
		return
	}

	err = updateApt(ctx)
	if err != nil {
		err = errors.Wrapf(err,
			"failed initial apt update")
		return
	}

	dir, err := ioutil.TempDir("", "estaleiro-deb-packages")
	if err != nil {
		err = errors.Wrapf(err,
			"failed creating temp directory for debian packages")
		return
	}

	defer os.RemoveAll(dir)

	err = os.Chmod(dir, 0755)
	if err != nil {
		err = errors.Wrapf(err,
			"failed changing permissions of temporary deb directory %s", dir)
		return
	}

	locations, err := gatherDebLocations(ctx, packages)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to install packages %v", packages)
		return
	}

	err = downloadDebianPackages(ctx, dir, locations)
	if err != nil {
		err = errors.Wrapf(err,
			"failed downloading packages %v", packages)
		return
	}

	pkgs, err = createPackages(ctx, dir, locations)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to create packages bom")
		return
	}

	err = createDebianRepositoryIndex(dir, pkgs)
	if err != nil {
		err = errors.Wrapf(err,
			"failed creating debian repository index")
		return
	}

	err = forceLocalSourcesList(ctx, dir)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to force apt to use local repositories")
		return
	}

	err = installDebianPackages(ctx, pkgs)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to install downloaded debian packages")
		return
	}

	err = removeAptLists()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to remove apt repository listing after installation")
		return
	}

	return
}

func createDebianRepositoryIndex(dir string, pkgs []Package) (err error) {
	logger.Info("create-deb-repository", lager.Data{"dir": dir})

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

func createPackages(ctx context.Context, dir string, locations []AptDebLocation) (pkgs []Package, err error) {
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

			ref := control.Name + "=" + control.Version
			sourceLocations, err := aptSourcePackageUris(ctx, ref)
			if err != nil {
				err = errors.Wrapf(err,
					"failed to retrieve source for package %s",
					ref)
				return
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
		out  bytes.Buffer
		sess = logger.Session("get-deb-pkg-info", lager.Data{"filename": filename})
	)

	sess.Info("start")
	defer sess.Info("finish")

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

// forceLocalSource ensures that we're only able to retrieve debian packages
// from the directory where we downloaded our stuff to so that no other
// repositories can provide those (which wouldn't be appropriately tracked).
//
// as this involves performing changes in the filesystem, this may lead to a
// system without proper configuration for `apt`.
//
// ps.: backup files/dirs are kept with `-backup` suffixed names.
//
func forceLocalSourcesList(ctx context.Context, repositoryDir string) (err error) {
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

	err = removeAptLists()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to remove apt repository listing")
		return
	}

	err = updateApt(ctx)
	if err != nil {
		err = errors.Wrapf(err,
			"couldn't update apt repositories")
		return
	}

	return
}

func removeAptLists() error {
	return os.RemoveAll("/var/lib/apt/lists")
}

func updateApt(ctx context.Context) (err error) {
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

func installDebianPackages(ctx context.Context, pkgs []Package) (err error) {
	args := []string{"install", "--no-install-recommends", "--no-install-suggests"}

	for _, pkg := range pkgs {
		args = append(args, pkg.Name+"="+pkg.Version)
	}

	out, err := exec.CommandContext(ctx, "apt", args...).CombinedOutput()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to install packages %v - %s", args[3:], string(out))
		return
	}

	return
}

func gatherDebLocations(ctx context.Context, packages []string) ([]AptDebLocation, error) {
	return aptUris(ctx, "install", packages)
}

func aptSourcePackageUris(ctx context.Context, packageName string) ([]AptDebLocation, error) {
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

func writeInitialList() (err error) {
	logger.Info("write-initial-list")

	const sourcesListFmt = `
# There are four official repositories that one
# can retrieve directly from ubuntu:
#
#
#                | FREE     | NON-FREE
#    ------------+----------+-----------
#      SUPPORTED | main     | restricted
#    -----------------------------------
#    UNSUPPORTED | universe | multiverse
#
deb http://archive.ubuntu.com/ubuntu/ %s main restricted
deb-src http://archive.ubuntu.com/ubuntu/ %s main restricted

deb http://archive.ubuntu.com/ubuntu/ %s-updates main restricted
deb-src http://archive.ubuntu.com/ubuntu/ %s-updates main restricted

deb http://archive.ubuntu.com/ubuntu/ %s-security main restricted
deb-src http://archive.ubuntu.com/ubuntu/ %s-security main restricted
`

	f, err := os.OpenFile("/etc/apt/sources.list", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to open `/etc/apt/sources.list` file to write initial apt sources list")
		return
	}
	defer f.Close()

	info, err := GatherOsRelease()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to detect codename to be used for initial sources.list")
		return
	}

	_, err = fmt.Fprintf(f, sourcesListFmt,
		info.Codename, info.Codename, info.Codename, info.Codename, info.Codename, info.Codename,
	)
	if err != nil {
		err = errors.Wrapf(err,
			"failed writing initial sources.list to `/etc/apt/sources.list`")
		return
	}

	return
}

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

	info = scanCodename(f)

	return
}

func scanCodename(reader io.Reader) (info OsRelease) {
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
	sess := logger.Session("download-debian-package", lager.Data{"name": location.Name})

	sess.Info("start")
	defer sess.Info("finish")

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
