DOCKERVER :=`cat VERSION`
.DEFAULT_GOAL := nabu
VERSION :=`cat VERSION`
   
nabu:
	cd cmd/nabu; \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 env go build -o nabu

docker:
	podman build  --tag="fils/nabu:$(VERSION)"  --file=./build/Dockerfile .

dockerpush:
	podman push localhost/fils/nabu:$(VERSION) fils/nabu:$(VERSION)
	podman push localhost/fils/nabu:$(VERSION) fils/nabu:latest

publish:
	docker tag fils/nabu:$(VERSION) fils/nabu:latest
	docker push fils/nabu:$(VERSION) ; \
	docker push fils/nabu:latest

releases: nabu docker dockerpush publish
