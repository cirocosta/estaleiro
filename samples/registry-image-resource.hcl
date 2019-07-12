step "build" {
  dockerfile = "./Dockerfile"
  target     = "release"
}

image "concourse/registry-resource" {
  base_image {
    name = "library/ubuntu"
    ref  = "bionic"
  }

  files "/opt/resource/" {
    copy "step" {
      paths = ["/assets"]

      bom {
        version = "${var.version}"
        source  = "https://github.com/concourse/registry-image-resource"
      }
    }
  }
}
