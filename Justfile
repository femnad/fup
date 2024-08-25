default:
    @just --choose

build-ubuntu:
    podman image build -t fup:ubuntu -f Dockerfile.ubuntu .

run-ubuntu: build-ubuntu
    podman container run --name ubuntu --rm -d fup:ubuntu

test:
    go test ./... -tags=unit

test-integration:
    go test ./... -tags=integration
