package dpkg_test

import (
	"bytes"
	"fmt"

	"github.com/cirocosta/estaleiro/dpkg"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dpkg", func() {

	Describe("Scan", func() {

		var (
			content string
			done    bool
			err     error
			pkg     dpkg.DebControl
			scanner dpkg.Scanner
		)

		JustBeforeEach(func() {
			scanner = dpkg.NewScanner(bytes.NewReader([]byte(content)))
			pkg, done, err = scanner.Scan()
		})

		Context("on empty reader", func() {

			BeforeEach(func() {
				content = ""
			})

			It("succeeds", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("is done", func() {
				Expect(done).To(BeTrue())
			})

		})

		Context("reader with malformed content", func() {

			BeforeEach(func() {
				content = "dsahiudash"
			})

			It("fails", func() {
				Expect(err).To(HaveOccurred())
			})

		})

		Context("with well-formed content", func() {

			BeforeEach(func() {
				content = sampleWellFormedPackage1
			})

			It("succeeds", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("has package fields correctly filled", func() {
				Expect(pkg).NotTo(BeNil())
				Expect(pkg).To(Equal(dpkg.DebControl{
					Name:    "software-properties-common",
					Version: "0.96.24.32.9",
				}))
			})
		})

		Context("with two packages in the same reader", func() {

			BeforeEach(func() {
				content = fmt.Sprintf(`%s

%s`, sampleWellFormedPackage2, sampleWellFormedPackage1)
			})

			It("succeeds", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("is not done", func() {
				Expect(done).NotTo(BeTrue())
			})

			It("retrieves the contents of just the first package", func() {
				Expect(pkg).NotTo(BeNil())
				Expect(pkg).To(Equal(dpkg.DebControl{
					Name:    "binutils-x86-64-linux-gnu",
					Version: "2.30-21ubuntu1~18.04.2",
				}))
			})

			Context("calling Scan one more time", func() {

				JustBeforeEach(func() {
					pkg, done, err = scanner.Scan()
				})

				It("succeeds", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("is now done", func() {
					Expect(done).To(BeTrue())
				})

				It("retrieves the contents of just the second package", func() {
					Expect(pkg).NotTo(BeNil())
					Expect(pkg).To(Equal(dpkg.DebControl{
						Name:    "software-properties-common",
						Version: "0.96.24.32.9",
					}))
				})

			})
		})

	})

})
