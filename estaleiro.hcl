# syntax = cirocosta/estaleiro-frontend:rc


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

  file "/usr/test-file" {
    from_tarball "archive.tgz" {
      path = "b"
    }
  }

  file "/usr/dummy-file" {
    from_tarball "archive.tgz" {
      path = "c"
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


# a dummy tarball just to test the functionality
#
tarball "archive.tgz" {
  source_file "b" {
    vcs "git" {
      ref        = "${estaleiro-commit}"
      repository = "https://github.com/cirocosta/estaleiro"
    }
  }

  source_file "c" {
    vcs "git" {
      ref        = "${estaleiro-commit}"
      repository = "https://github.com/cirocosta/estaleiro"
    }
  }
}
