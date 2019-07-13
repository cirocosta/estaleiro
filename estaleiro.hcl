step "build" {
  dockerfile = "./Dockerfile"
  target     = "build"
}

image "cirocosta/estaleiro" {

  base_image {
    name = "ubuntu"
    ref  = "bionic"
  }

  file "/usr/local/bin/estaleiro" {
    from_step "build" {
      path = "/usr/local/bin/estaleiro"
    }
  }
}
