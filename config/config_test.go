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
		Entry("minimal", Case{
			content: `
			image "something" {
			  base_image {
			    name = "ubuntu"
			  }

			  env = {
			    "FOO": "bar",
			  }

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
					Env:        map[string]string{"FOO": "bar"},
				},
			},
		}),
		Entry("with step", Case{
			content: `
			step "build" {
			  dockerfile = "./Dockerfile"
			}

			image "something" {
			  base_image {
			    name = "ubuntu"
			  }

			  file "/usr/local/bin/estaleiro" {
			    from_step "build" {
			      path = "/bin/estaleiro"
			    }
			  }
			}
			`,
			expected: config.Config{
				Image: config.Image{
					Name: "something",
					BaseImage: config.BaseImage{
						Name: "ubuntu",
					},
					Files: []config.File{
						{
							Destination: "/usr/local/bin/estaleiro",
							FromStep: &config.FileFromStep{
								StepName: "build",
								Path:     "/bin/estaleiro",
							},
						},
					},
				},
				Steps: []config.Step{
					{
						Name:       "build",
						Dockerfile: "./Dockerfile",
					},
				},
			},
		}),
		Entry("with tarball", Case{
			content: `
			tarball "linux-rc" {
			  source_file "concourse/bin/gdn" {
			    vcs "git" {
			      ref        = "master"
			      repository = "https://github.com/cloudfoundry/guardian"
			    }
			  }
			}

			image "something" {
			  base_image {
			    name = "ubuntu"
			  }

			  file "/usr/local/concourse/bin/gdn" {
			    from_tarball "linux-rc" {
			      path = "concourse/bin/gdn"
			    }
			  }
			}
			`,
			expected: config.Config{
				Image: config.Image{
					Name: "something",
					BaseImage: config.BaseImage{
						Name: "ubuntu",
					},
					Files: []config.File{
						{
							Destination: "/usr/local/concourse/bin/gdn",
							FromTarball: &config.FileFromTarball{
								TarballName: "linux-rc",
								Path:        "concourse/bin/gdn",
							},
						},
					},
				},
				Tarballs: []config.Tarball{
					{
						Name: "linux-rc",
						SourceFiles: []config.SourceFile{
							{
								Location: "concourse/bin/gdn",
								VCS: config.VCS{
									Type:       "git",
									Ref:        "master",
									Repository: "https://github.com/cloudfoundry/guardian",
								},
							},
						},
					},
				},
			},
		}),
	)
})
