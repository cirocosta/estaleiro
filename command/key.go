package command

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"os/exec"

	"code.cloudfoundry.org/lager"
	bomfs "github.com/cirocosta/estaleiro/bom/fs"
	"github.com/cirocosta/estaleiro/dpkg"
	"github.com/pkg/errors"
	gpg "golang.org/x/crypto/openpgp"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

type keyCommand struct {
	Output string `long:"output"`
}

func (c *keyCommand) Execute(args []string) (err error) {
	ctx := context.TODO()

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

	logger.Info("download-keys", lager.Data{"keys": args})
	dests, err := downloadKeys(ctx, args)
	if err != nil {
		return
	}

	logger.Info("add-keys", lager.Data{"dests": dests})
	err = addKeys(ctx, dests)
	if err != nil {
		return
	}

	keys, err := readKeyRing()
	if err != nil {
		return
	}

	w, err := writer(c.Output)
	if err != nil {
		return
	}

	res, err := yaml.Marshal(bomfs.NewKeysV1(false, keys))
	if err != nil {
		return
	}

	_, err = fmt.Fprintf(w, "%s", string(res))
	if err != nil {
		err = errors.Wrapf(err,
			"failed writing bill of materials to %s", c.Output)
		return
	}

	return
}

func readKeyRing() (keys []bomfs.Key, err error) {
	const fname = "/etc/apt/trusted.gpg"

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
