package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cirocosta/estaleiro/config"
)

var _ = Describe("Config", func() {

	Describe("Parse", func() {

		const mockFilename = "mock-file"

		var (
			content string
			err     error
		)

		JustBeforeEach(func() {
			_, err = config.Parse([]byte(content), mockFilename)
		})

		Context("with empty content", func() {
			It("fails", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("having image", func() {
			Context("missing `base_image`", func() {
				BeforeEach(func() {
					content = `image "busybox" { }`
				})

				It("fails", func() {
					Expect(err).To(HaveOccurred())
				})
			})

			Context("having `base_image`", func() {
				BeforeEach(func() {
					content = `image "busybox" { 
						base_image {
							name = "this"	
						}
					}`
				})

				It("succeeds", func() {
					Expect(err).ToNot(HaveOccurred())
				})
			})

		})

	})

})
