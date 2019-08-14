# syntax=cirocosta/estaleiro

image "concourse/s3-resource" {
  base_image = "ubuntu:bionic"

  apt {
    package "tzdata" {}
    package "ca-certificates" {}
    package "unzip" {}
    package "zip" {}
  }

  file "/opt/resource/check" {
    from_step "build" {
      path = "/assets/check"
    }
  }

  file "/opt/resource/in" {
    from_step "build" {
      path = "/assets/in"
    }
  }

  file "/opt/resource/out" {
    from_step "build" {
      path = "/assets/out"
    }
  }

}

step "build" {
  dockerfile = "./s3-resource/dockerfiles/ubuntu/Dockerfile"
  context    = "./s3-resource"
  target     = "builder"

  source_file "/assets/*" {
    vcs "git" {
      repository = "https://github.com/concourse/s3-resource"
      ref        = "${ref}"
    }
  }
}
