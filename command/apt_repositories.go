package command

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/lager"
	bomfs "github.com/cirocosta/estaleiro/bom/fs"
	"github.com/cirocosta/estaleiro/dpkg"
	"github.com/pkg/errors"
	gpg "golang.org/x/crypto/openpgp"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

// retrieves the necessary repository information in order to be able to
// retrieve package information.
//
// - /var/lib/apt/lists
// - /etc/apt/trusted.gpg
//
//
type aptRepositoriesCommand struct {
	Output       string   `long:"output" default:"-"`
	Repositories []string `short:"r"`
	Keys         []string `short:"k"`
}

func (c *aptRepositoriesCommand) Execute(args []string) (err error) {
	ctx := context.TODO()

	keys, err := c.setupRepositoriesAndKeys(ctx)
	if err != nil {
		return
	}

	err = writeKeysBom(c.Output, keys)
	if err != nil {
		return
	}

	return
}

func writeKeysBom(fname string, keys []bomfs.Key) (err error) {
	w, err := writer(fname)
	if err != nil {
		err = errors.Wrapf(err,
			"failed creating writer for file %s", fname)
		return
	}

	b, err := yaml.Marshal(bomfs.NewKeysV1(false, keys))
	if err != nil {
		err = errors.Wrapf(err,
			"failed marshalling bomfs keys")
		return
	}

	_, err = fmt.Fprintf(w, "%s", string(b))
	if err != nil {
		err = errors.Wrapf(err,
			"failed writing bill of materials to %s", fname)
		return
	}

	return
}

func fetchAndImportKeys(ctx context.Context, uris []string) (err error) {
	logger.Info("download-keys", lager.Data{"uris": uris})
	dests, err := downloadKeys(ctx, uris)
	if err != nil {
		return
	}

	logger.Info("add-keys", lager.Data{"dests": dests})
	err = addKeys(ctx, dests)
	if err != nil {
		return
	}

	return
}

func (c *aptRepositoriesCommand) setupRepositoriesAndKeys(ctx context.Context) (keys []bomfs.Key, err error) {
	logger.Info("update-apt")
	err = dpkg.UpdateApt(ctx)
	if err != nil {
		return
	}

	logger.Info("install-deps")
	err = installDependencies(ctx)
	if err != nil {
		return
	}

	if len(c.Keys) > 0 {
		err = fetchAndImportKeys(ctx, c.Keys)
		if err != nil {
			return
		}
	}

	logger.Info("reading-keyring")
	keys, err = readKeyRings()
	if err != nil {
		return
	}

	logger.Info("remove-apt-lists")
	err = dpkg.RemoveAptLists()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to remove apt repository listing after installation")
		return
	}

	supportedRepos, err := dpkg.UbuntuSupportedRepositories()
	if err != nil {
		err = errors.Wrapf(err,
			"failed generating list of ubuntu supported repos")
		return
	}

	repos := append(c.Repositories, supportedRepos...)

	logger.Info("updating-apt-sources-list", lager.Data{"repos": repos})
	err = dpkg.WriteSourcesList(repos)
	if err != nil {
		err = errors.Wrapf(err,
			"failed setting up initial sources.list")
		return
	}

	err = dpkg.UpdateApt(ctx)
	if err != nil {
		err = errors.Wrapf(err, "failed apt update")
		return
	}

	return
}

func readKeyRingFile(fname string) (keys []bomfs.Key, err error) {
	f, err := os.Open(fname)
	if err != nil {
		err = errors.Wrapf(err, "failed opening %s", fname)
		return
	}

	defer f.Close()

	list, err := gpg.ReadKeyRing(f)
	if err != nil {
		err = errors.Wrapf(err, "failed reading key ring from %s", fname)
		return
	}

	keys = make([]bomfs.Key, len(list))
	for idx, entity := range list {
		keys[idx].Fingerprint = hex.EncodeToString(entity.PrimaryKey.Fingerprint[:])
		keys[idx].Identities = make([]string, len(entity.Identities))

		k := 0
		for name := range entity.Identities {
			keys[idx].Identities[k] = name
			k++
		}
	}

	return
}

func tryReadLocalTrustedKeys() (keys []bomfs.Key, err error) {
	const fname = "/etc/apt/trusted.gpg"

	_, err = os.Stat(fname)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}

		err = errors.Wrapf(err,
			"failed checking existence of local trusted keys file %s", fname)
		return
	}

	keys, err = readKeyRingFile(fname)
	return
}

func readKeyRings() (keys []bomfs.Key, err error) {
	keys, err = tryReadLocalTrustedKeys()
	if err != nil {
		err = errors.Wrapf(err,
			"failed trying to read local trusted keys file")
		return
	}

	matches, err := filepath.Glob("/etc/apt/trusted.gpg.d/*.gpg")
	if err != nil {
		err = errors.Wrapf(err,
			"failed creating flob for trusted gpg directory files")
		return
	}

	var newKeys []bomfs.Key
	for _, match := range matches {
		newKeys, err = readKeyRingFile(match)
		if err != nil {
			err = errors.Wrapf(err,
				"failed reading keyring from trusted keyrings dir")
			return
		}

		keys = append(keys, newKeys...)
	}

	return
}

func randomName(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"

	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}

	return string(b)
}

func downloadKeys(ctx context.Context, uris []string) (dests []string, err error) {
	var eg *errgroup.Group

	eg, ctx = errgroup.WithContext(ctx)
	dests = make([]string, len(uris))

	for idx, uri := range uris {
		uri := uri
		fname := randomName(10)

		dests[idx] = fname
		eg.Go(func() error { return dpkg.Download(ctx, uri, fname) })
	}

	err = eg.Wait()
	if err != nil {
		err = errors.Wrapf(err,
			"failed retrieving public keys")
		return
	}

	return
}

func installDependencies(ctx context.Context) (err error) {
	args := []string{"install", "--yes", "--no-install-recommends", "--no-install-suggests",
		"gnupg", "ca-certificates",
	}

	out, err := exec.CommandContext(ctx, "apt", args...).CombinedOutput()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to apt dependencies %v - %s",
			args[4:], string(out))
		return
	}

	return
}

func addToKeyRing(ctx context.Context, key string) (err error) {
	out, err := exec.CommandContext(ctx, "apt-key", "add", key).CombinedOutput()
	if err != nil {
		err = errors.Wrapf(err,
			"failed adding key %s to keyring - %s",
			key, string(out))
		return
	}

	return
}

func addKeys(ctx context.Context, keys []string) (err error) {
	for _, key := range keys {
		err = addToKeyRing(ctx, key)
		if err != nil {
			return
		}
	}

	return
}
