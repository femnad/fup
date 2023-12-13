terraform {
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "3.0.2"
    }
  }
}

provider "docker" {
  registry_auth {
    address     = "registry-1.docker.io"
    config_file = "~/.docker/config.json"
  }
}
