locals {
  registry          = "registry-1.docker.io"
  fedora_tag        = "fedora-ci"
  fedora_image_name = "${local.registry}/femnad/fup:${local.fedora_tag}"
  fedora_dockerfile = "Dockerfile.fedora-ci"
  ubuntu_tag        = "ubuntu-ci"
  ubuntu_image_name = "${local.registry}/femnad/fup:${local.ubuntu_tag}"
  ubuntu_dockerfile = "Dockerfile.ubuntu-ci"
}

resource "docker_image" "fedora" {
  name = local.fedora_image_name
  build {
    context    = "."
    dockerfile = local.fedora_dockerfile
    tag        = [local.fedora_tag]
  }

  triggers = {
    dir_sha1 = sha1(filesha1(local.fedora_dockerfile))
  }
}

resource "docker_registry_image" "fedora" {
  name          = docker_image.fedora.name
  keep_remotely = true
}

resource "docker_image" "ubuntu" {
  name = local.ubuntu_image_name
  build {
    context    = "."
    dockerfile = local.ubuntu_dockerfile
    tag        = [local.ubuntu_tag]
  }

  triggers = {
    dir_sha1 = sha1(filesha1(local.ubuntu_dockerfile))
  }
}

resource "docker_registry_image" "ubuntu" {
  name          = docker_image.ubuntu.name
  keep_remotely = true
}
