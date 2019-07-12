image "concourse/concourse" {

  description = <<DESC
    The container image used to distribute Concourse for Containerized
    platforms (like Docker and Kubernetes), having "all bateries" included":
    - resource types
    - runtime binaries
    - fly
    - concourse.
DESC

  base_image {
    name = "library/ubuntu"
    ref = "bionic"
  }

  package {
    "btrfs-tools" { }
    "ca-certificates" {}
    "dumb-init" {}
    "iproute2" {}
    "file" {}
  }

  env {
    CONCOURSE_SESSION_SIGNING_KEY = "/concourse-keys/session_signing_key"
  }

  files "/usr/local/concourse/" {
    description = "the Concourse binaries and resource types exist"

    from {
      local_tarball "./linux-rc/*.tgz" {
        description = ""
        bill_of_materials = "bom.yml"
      }
    }
  }

  entrypoint = ["/usr/local/concourse/bin/concourse"]

  stopsignal = "SIGUSR2"
}
