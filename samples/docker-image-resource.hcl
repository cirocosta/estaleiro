# syntax=cirocosta/estaleiro

image "concourse/docker-image-resource" {
  base_image {
    name = "ubuntu"
    ref  = "bionic"
  }

  apt {
    package "ca-certificates" {}
  }

  apt {
    repositories = [
      "deb http://archive.ubuntu.com/ubuntu/ bionic universe",
      "deb-src http://archive.ubuntu.com/ubuntu/ bionic universe",
      "deb https://download.docker.com/linux/ubuntu bionic stable",
    ]

    key "docker" {
      uri = "https://download.docker.com/linux/ubuntu/gpg"
    }

    package "jq" {}
    package "docker-ce" { }
  }
}
