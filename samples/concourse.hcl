image "concourse/concourse" {
  base_image {
    name = "library/ubuntu"
    ref  = "bionic"
  }

  package {
    "btrfs-tools" {}
    "ca-certificates" {}
    "dumb-init" {}
    "file" {}
    "iproute2" {}
  }

  env {
    CONCOURSE_SESSION_SIGNING_KEY = "/concourse-keys/session_signing_key"
  }

  files "/usr/local/concourse/" {
    description = "the Concourse binaries and resource types exist"

    copy "local" {
      paths = "./linux-rc/concourse-*.tgz"
    }
  }

  entrypoint = ["/usr/local/concourse/bin/concourse"]

  stopsignal = "SIGUSR2"
}
