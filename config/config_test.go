package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/cirocosta/estaleiro/config"
)

var _ = Describe("Config", func() {

	type Case struct {
		content    string
		vars       map[string]string
		shouldFail bool
		expected   config.Config
	}

	const mockFilename = "mock-file"

	DescribeTable("Parse",
		func(c Case) {
			cfg, err := config.Parse([]byte(c.content), mockFilename, c.vars)

			if c.shouldFail {
				Expect(err).To(HaveOccurred())
				return
			}

			Expect(err).NotTo(HaveOccurred())
			Expect(*cfg).To(Equal(c.expected))
		},
		Entry("empty content", Case{
			shouldFail: true,
		}),

		Entry("success", Case{
			content: `
			step "build" {
			  dockerfile = "./Dockerfile"
			}

			image "something" {
			  base_image {
			    name = "ubuntu"
			  }

			  env = [
			    "FOO=bar",
			  ]

			  entrypoint = ["/bin/bash"]
			}
			`,
			expected: config.Config{
				Image: config.Image{
					Name: "something",
					BaseImage: config.BaseImage{
						Name: "ubuntu",
					},
					Entrypoint: []string{"/bin/bash"},
				},
				Steps: []config.Step{
					{
						Name:       "build",
						Dockerfile: "./Dockerfile",
					},
				},
			},
		}),
	)
})
