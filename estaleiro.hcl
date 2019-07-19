# syntax = cirocosta/estaleiro-frontend:rc

step "build" {
  dockerfile = "./Dockerfile"
  target     = "build"

  file "/bin/estaleiro" {
    source "git" {
      ref  = "master"
      root = "https://github.com/cirocosta/estaleiro"
    }
  }
}

image "cirocosta/estaleiro" {

  base_image {
    name = "ubuntu"
    ref  = "bionic"
  }

  file "/usr/local/bin/estaleiro" {
    from_step "build" "/usr/local/bin/estaleiro" { }
  }
}
