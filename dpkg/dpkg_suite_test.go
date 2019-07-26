package dpkg_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const sampleWellFormedPackage1 = `Package: software-properties-common
Status: install ok installed
Priority: optional
Section: admin
Installed-Size: 196
Maintainer: Michael Vogt <michael.vogt@ubuntu.com>
Architecture: all
Source: software-properties
Version: 0.96.24.32.9
Replaces: python-software-properties (<< 0.85), python3-software-properties (<< 0.85)
Depends: python3:any (>= 3.3.2-2~), python3, python3-gi, gir1.2-glib-2.0, python-apt-common (>= 0.9), python3-dbus, python3-software-properties (= 0.96.24.32.9), ca-certificates
Breaks: python-software-properties (<< 0.85), python3-software-properties (<< 0.85)
Conffiles:
 /etc/dbus-1/system.d/com.ubuntu.SoftwareProperties.conf cc3c01a5b5e8e05d40c9c075f44c43ea
Description: manage the repositories that you install software from (common)
 This software provides an abstraction of the used apt repositories.
 It allows you to easily manage your distribution and independent software
 vendor software sources.
 .
 This package contains the common files for software-properties like the
 D-Bus backend.`

const sampleWellFormedPackage2 = `Package: binutils-x86-64-linux-gnu
Status: install ok installed
Priority: optional
Section: devel
Installed-Size: 11758
Maintainer: Ubuntu Core developers <ubuntu-devel-discuss@lists.ubuntu.com>
Architecture: amd64
Multi-Arch: foreign
Source: binutils
Version: 2.30-21ubuntu1~18.04.2
Replaces: binutils (<< 2.29-6)
Depends: binutils-common (= 2.30-21ubuntu1~18.04.2), libbinutils (= 2.30-21ubuntu1~18.04.2), libc6 (>= 2.27), zlib1g (>= 1:1.1.4)
Suggests: binutils-doc (= 2.30-21ubuntu1~18.04.2)
Breaks: binutils (<< 2.29-6)
Description: GNU binary utilities, for x86-64-linux-gnu target
 This package provides GNU assembler, linker and binary utilities
 for the x86-64-linux-gnu target.
 .
 You don't need this package unless you plan to cross-compile programs
 for x86-64-linux-gnu and x86-64-linux-gnu is not your native platform.
Homepage: https://www.gnu.org/software/binutils/
Original-Maintainer: Matthias Klose <doko@debian.org>`

func TestDpkg(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dpkg Suite")
}
