docker-build:
	docker build -t caarlos0/domain_exporter .

docker-push:
	docker push caarlos0/domain_exporter

docker: docker-build docker-push

test:
	go test -v ./...

setup:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure -v
