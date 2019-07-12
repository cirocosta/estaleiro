image "concourse/git-resource" {

  base_image {
    name = "library/ubuntu"
    ref  = "bionic"
  }

  repository "backports" {
    uri = "deb http://http.debian.net/debian wheezy-backports main"
  }

  package {
    curl {}
    gnupg {}
    gzip {}

    git {
      without_files = [
        "git-add"
        "git-rm"
      ]
    }
  }

}
