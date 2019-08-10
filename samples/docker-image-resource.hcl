# syntax=cirocosta/estaleiro

image "concourse/docker-image-resource" {
  base_image {
    name = "ubuntu"
    ref  = "bionic"
  }

  apt {
    package "ca-certificates" {}
  }

  # non-supported packages
  #
  apt {
    repositories = [
      "deb http://archive.ubuntu.com/ubuntu/ bionic universe",
      "deb-src http://archive.ubuntu.com/ubuntu/ bionic universe",
    ]

    package "jq" {}
  }

  apt {
    repositories = [
      "deb http://archive.ubuntu.com/ubuntu/ bionic universe",
      "deb-src http://archive.ubuntu.com/ubuntu/ bionic universe",
    ]

    key "docker" {
      uri = "https://download.docker.com/linux/ubuntu/gpg"
    }
  }
}
