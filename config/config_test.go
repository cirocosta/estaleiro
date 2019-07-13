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

	})

})
