package dpkg_test

import (
	"bytes"

	"github.com/cirocosta/estaleiro/dpkg"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Apt", func() {

	Describe("ScanAptUris", func() {

		var (
			err     error
			content string
			res     []dpkg.AptDebLocation
		)

		JustBeforeEach(func() {
			reader := bytes.NewReader([]byte(content))
			res, err = dpkg.ScanAptDebLocations(reader)
		})

		Context("with no content", func() {
			BeforeEach(func() {
				content = ""
			})

			It("succeeds", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns 0-length array", func() {
				Expect(res).To(HaveLen(0))
			})
		})

		Context("with proper content", func() {

			BeforeEach(func() {
				content = sampleAptPrintUris
			})

			It("succeeds", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("properly parses it", func() {
				Expect(res).To(ConsistOf([]dpkg.AptDebLocation{
					{
						URI:    `http://archive.ubuntu.com/ubuntu/pool/main/p/perl/perl-modules-5.26_5.26.1-6ubuntu0.3_all.deb`,
						Name:   `perl-modules-5.26_5.26.1-6ubuntu0.3_all.deb`,
						Size:   `2762592`,
						MD5sum: `e3bb462a24dda2bed9eeb0136b8d0b87`,
					},
					{
						URI:    `http://archive.ubuntu.com/ubuntu/pool/main/g/gdbm/libgdbm5_1.14.1-6_amd64.deb`,
						Name:   `libgdbm5_1.14.1-6_amd64.deb`,
						Size:   `26040`,
						MD5sum: `99523c29e5ed1272dff7abc066eec3f9`,
					},
				}))
			})
		})
	})

})

const sampleAptPrintUris = `Reading package lists...
Building dependency tree...
Reading state information...
The following additional packages will be installed:
  libgdbm-compat4 libgdbm5 libperl5.26 netbase perl-modules-5.26
Suggested packages:
  gdbm-l10n perl-doc libterm-readline-gnu-perl | libterm-readline-perl-perl
  make
The following NEW packages will be installed:
  libgdbm-compat4 libgdbm5 libperl5.26 netbase perl perl-modules-5.26
0 upgraded, 6 newly installed, 0 to remove and 0 not upgraded.
Need to get 6536 kB of archives.
After this operation, 41.6 MB of additional disk space will be used.
'http://archive.ubuntu.com/ubuntu/pool/main/p/perl/perl-modules-5.26_5.26.1-6ubuntu0.3_all.deb' perl-modules-5.26_5.26.1-6ubuntu0.3_all.deb 2762592 MD5Sum:e3bb462a24dda2bed9eeb0136b8d0b87
'http://archive.ubuntu.com/ubuntu/pool/main/g/gdbm/libgdbm5_1.14.1-6_amd64.deb' libgdbm5_1.14.1-6_amd64.deb 26040 MD5Sum:99523c29e5ed1272dff7abc066eec3f9`
