.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix cgo -o plugin ./cmd/plugin

.PHONY: docker
docker:
	docker build -t zc2638/drone-k8s-plugin -f build/Dockerfile .
