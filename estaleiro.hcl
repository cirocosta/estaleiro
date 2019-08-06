# syntax = cirocosta/estaleiro


# the final image to produce
#
image "cirocosta/estaleiro" {
  base_image {
    name = "ubuntu"
    ref  = "bionic"
  }

  apt {
    package "ca-certificates" {}
  }

  file "/usr/local/bin/estaleiro" {
    from_step "build" {
      path = "/bin/estaleiro"
    }
  }

  entrypoint = ["/usr/local/bin/estaleiro", "frontend"]
}


# performs the build of `estaleiro`.
#
step "build" {
  dockerfile = "./Dockerfile"
  target     = "build"

  source_file "/bin/estaleiro" {
    vcs "git" {
      ref        = "${estaleiro-commit}"
      repository = "https://github.com/cirocosta/estaleiro"
    }
  }
}

