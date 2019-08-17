# syntax=cirocosta/estaleiro

image "concourse/docker-image-resource" {
  base_image = "ubuntu:bionic"

  apt {
    repositories = [
      "deb http://archive.ubuntu.com/ubuntu/ bionic universe",
      "deb-src http://archive.ubuntu.com/ubuntu/ bionic universe",
      "deb https://download.docker.com/linux/ubuntu bionic stable",
    ]

    key "docker" {
      uri = "https://download.docker.com/linux/ubuntu/gpg"
    }

    package "ca-certificates" {}
    package "jq" {}

    package "docker-ce" {
      version = "5:19.03.1~3-0~ubuntu-bionic"

      vcs "git" {
        repository = "https://github.com/docker/docker-ce"
        ref        = "19.03.1"
      }
    }

    package "containerd.io" {
      version = "1.2.6-3"

      vcs "git" {
        repository = "https://github.com/containerd/containerd"
        ref        = "v1.2.6"
      }
    }

    package "docker-ce-cli" {
      version = "5:19.03.1~3-0~ubuntu-bionic"

      vcs "git" {
        repository = "https://github.com/docker/cli"
        ref        = "v19.03.1 "
      }
    }
  }
}
