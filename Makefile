BINARY := gleaner
VERSION :=`cat VERSION`
.DEFAULT_GOAL := gleaner

gleaner:
	cd cmd/$(BINARY) ; \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 env go build -o $(BINARY)

gleaner.exe:
	cd cmd/$(BINARY) ; \
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 env go build -o $(BINARY).exe

gleaner.darwin:
	cd cmd/$(BINARY) ; \
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 env go build -o $(BINARY)_darwin

releases: gleaner gleaner.exe gleaner.darwin

docker:
	docker build  --tag="nsfearthcube/gleaner:$(VERSION)"  --file=./build/Dockerfile . ; \
	docker tag nsfearthcube/gleaner:$(VERSION) nsfearthcube/gleaner:latest

removeimage:
	docker rmi --force nsfearthcube/gleaner:$(VERSION)
	docker rmi --force nsfearthcube/gleaner:latest

publish: docker
	docker push nsfearthcube/gleaner:$(VERSION) ; \
	docker push nsfearthcube/gleaner:latest

glweb:
	cd cmd/glweb ; \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 env go build -o glweb

dockerweb:
	docker build  --tag="fils/gleanerweb:$(VERSION)"  --file=./build/DockerfileWeb . ; \
	docker tag fils/gleanerweb:$(VERSION) fils/gleanerweb:latest

publishweb:
	docker push fils/gleanerweb:$(VERSION) ; \
	docker push fils/gleanerweb:latest
