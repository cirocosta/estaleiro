# syntax=cirocosta/estaleiro

image "concourse/s3-resource" {
  base_image {
    name = "ubuntu"
    ref  = "bionic"
  }

  apt {

    repositories = [
      "deb http://archive.ubuntu.com/ubuntu/ bionic universe",
      "deb-src http://archive.ubuntu.com/ubuntu/ bionic universe",
    ]

    package "ca-certificates" {}
    package "g++" {}
    package "git" {}
    package "git-crypt" {}
    package "git-lfs" {}
    package "gnupg" {}
    package "gzip" {}
    package "jq" {}
    package "libssl-dev" {}
    package "libstdc++6" {}
    package "make" {}
    package "openssh-client" {}
    package "openssl" {}
    package "proxytunnel" {}
  }


  # the files needed for implementing the resource
  # type interface.
  #

  file "/opt/resource/check" {
    from_tarball "assets.tgz" {
      path = "check"
    }
  }

  file "/opt/resource/out" {
    from_tarball "assets.tgz" {
      path = "out"
    }
  }

  file "/opt/resource/in" {
    from_tarball "assets.tgz" {
      path = "in"
    }
  }


  ## helper files.
  ##

  file "/opt/resource/askpass.sh" {
    from_tarball "assets.tgz" {
      path = "askpass.sh"
    }
  }

  file "/opt/resource/common.sh" {
    from_tarball "assets.tgz" {
      path = "common.sh"
    }
  }


  file "/opt/resource/deepen_shallow_clone_until_ref_is_found_then_check_out" {
    from_tarball "assets.tgz" {
      path = "deepen_shallow_clone_until_ref_is_found_then_check_out"
    }
  }
}


tarball "assets.tgz" {
  source_file "*" {
    vcs "git" {
      repository = "https://github.com/concourse/s3-resource"
      ref        = "${ref}"
    }
  }
}
