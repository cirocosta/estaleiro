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
			cfg     *config.Config
		)

		JustBeforeEach(func() {
			cfg, err = config.Parse([]byte(content), mockFilename)
		})

		Context("with empty content", func() {
			It("fails", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with just step", func() {
			BeforeEach(func() {
				content = `step "build" { }`
			})

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

			Context("having `step`", func() {

				Context("without file defined", func() {
					BeforeEach(func() {
						content = `
					step "this" {
						dockerfile = "./Dockerfile"
					}

					image "busybox" { 
						base_image {
							name = "this"	
						}
					}`
					})

					It("fails", func() {
						Expect(err).To(HaveOccurred())
					})
				})

				Context("with file defined", func() {

					BeforeEach(func() {
						content = `
					step "this" {
					  dockerfile = "./Dockerfile"

					  file "/bin/binary" {
					    vcs "git" {
					       repository = "something.git"
					       ref        = "master"
					    }
					  }
					}

					image "busybox" { 
					  base_image {
					    name = "this"	
					  }
					}`
					})

					It("succeeds", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(cfg.Steps[0].File.Location).To(Equal("/bin/binary"))
					})

				})

			})
		})

	})

})
