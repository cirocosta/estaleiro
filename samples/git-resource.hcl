# syntax=cirocosta/estaleiro

image "concourse/git-resource" {
  base_image = "ubuntu:bionic"

  apt {
    package "ca-certificates" {}
    package "g++" {}
    package "git" {}
    package "gnupg" {}
    package "gzip" {}
    package "libssl-dev" {}
    package "libstdc++6" {}
    package "make" {}
    package "openssh-client" {}
    package "openssl" {}
  }

  # free but unsupported packages
  #

  apt {

    repositories = [
      "deb http://archive.ubuntu.com/ubuntu/ bionic universe",
      "deb-src http://archive.ubuntu.com/ubuntu/ bionic universe",
    ]

    package "git-crypt" {}
    package "git-lfs" {}
    package "jq" {}
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
