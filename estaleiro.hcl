# syntax = cirocosta/estaleiro-frontend:rc

# a step to be build so that we can consume files from it.
#
step "build" {
  dockerfile = "./Dockerfile"
  target     = "build"

  # declaring a file that can be consumed by the final image
  #
  source_file "/usr/local/bin/estaleiro" {
    vcs "git" {
      ref        = "${estaleiro-commit}"
      repository = "https://github.com/cirocosta/estaleiro"
    }
  }
}

# the final image to produce
#
image "cirocosta/estaleiro" {
  base_image {
    name = "ubuntu"
    ref  = "bionic"
  }

  # retrieving the file from the step 
  #
  file "/usr/local/bin/estaleiro" {
    from_step "build" {
      path = "/usr/local/bin/estaleiro"
    }
  }
}
