rule "base_image" {
  name "match" {
    description = "must come from `ubuntu`"
    value       = "ubuntu"
  }

  ref "match" {
    description = "must have been tagged as bionic"
    value       = "bionic"
  }

  ref "match" {
    description = "must have a digest provided"
    value       = ""
  }
}

rule "file" {
  bill_of_materials {
    required = true
  }
}

rule "package" {
  source {
    required = true
  }
}

rule "package" "gcc" {
  source {
    required = false
  }
}
