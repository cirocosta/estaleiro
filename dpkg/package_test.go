package dpkg_test

import (
	"github.com/cirocosta/estaleiro/dpkg"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("debcontrol", func() {

	Describe("ControlString", func() {

		It("does the right thing", func() {
			res := (dpkg.DebControl{
				Name:          "software-properties-common",
				Version:       "0.96.24.32.9",
				SourcePackage: "software-properties",
				Architecture:  "all",
				Maintainer:    "Michael Vogt <michael.vogt@ubuntu.com>",
				Description:   "manage the repositories that you install software from (common)",
			}).ControlString()
			Expect(res).To(Equal(`Package: software-properties-common
Source: software-properties
Architecture: all
Description: manage the repositories that you install software from (common)
Maintainer: Michael Vogt <michael.vogt@ubuntu.com>
Version: 0.96.24.32.9

`))
		})
	})

})
