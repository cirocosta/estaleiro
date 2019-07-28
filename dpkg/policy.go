package dpkg

// TODO this can potentially be discarded

// import (
// 	"fmt"
// 	"os"
// 	"path"
// 	"path/filepath"
// 	"strings"

// 	"github.com/pkg/errors"
// )

// type AptRepository struct {
// 	RepositoryURL string
// }

// func AptFileToRepository(filename string) (repo string) {
// 	repo = strings.ReplaceAll(path.Base(filename), "_", "/")
// 	return
// }

// const packagesGlob = "*_Packages"

// // find /var/lib/apt/lists/ -name "*_Packages"
// func LoadRepositoriesPackages(directory string) (listing map[string][]Package, err error) {
// 	fullGlob := filepath.Join(directory, packagesGlob)

// 	matches, err := filepath.Glob(fullGlob)
// 	if err != nil {
// 		err = errors.Wrapf(err,
// 			"failed preparing glob for deb Packages file search")
// 		return
// 	}

// 	listing = make(map[string][]Package, len(matches))
// 	for _, match := range matches {
// 		var (
// 			repository = AptFileToRepository(match)
// 			pkgs       []Package
// 		)

// 		pkgs, err = scanPackagesFile(match)
// 		if err != nil {
// 			err = errors.Wrapf(err,
// 				"failed scaning repositories for repository %s", repository)
// 			return
// 		}

// 		listing[repository] = pkgs
// 	}

// 	fmt.Println(listing)

// 	return
// }

// func scanPackagesFile(filename string) (pkgs []Package, err error) {
// 	f, err := os.Open(filename)
// 	if err != nil {
// 		err = errors.Wrapf(err,
// 			"failed opening package file %s", filename)
// 		return
// 	}

// 	defer f.Close()

// 	scanner := NewScanner(f)
// 	return scanner.ScanAll()
// }
