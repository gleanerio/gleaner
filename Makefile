BINARY := gleaner
BINARYIO := glcon
VERSION :=`cat VERSION`
.DEFAULT_GOAL := gleaner

gleaner:
	cd cmd/$(BINARY) ; \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 env go build -ldflags "-X main.VERSION=$(VERSION)" -o $(BINARY)

gleaner.exe:
	cd cmd/$(BINARY) ; \
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 env go build -ldflags "-X main.VERSION=$(VERSION)" -o $(BINARY).exe

gleaner.darwin:
	cd cmd/$(BINARY) ; \
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 env go build  -ldflags "-X main.VERSION=$(VERSION)" -o $(BINARY)_darwin

glcon:
	cd cmd/$(BINARYIO) ; \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 env go build -ldflags "-X main.VERSION=$(VERSION)" -o $(BINARYIO)

glcon.exe:
	cd cmd/$(BINARYIO) ; \
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 env go build -ldflags "-X main.VERSION=$(VERSION)" -o $(BINARYIO).exe

glcon.darwin:
	cd cmd/$(BINARYIO) ; \
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 env go build -ldflags "-X main.VERSION=$(VERSION)" -o $(BINARYIO)_darwin

releases: gleaner gleaner.exe gleaner.darwin glcon glcon.exe glcon.darwin

docker:
	podman build  --tag="nsfearthcube/gleaner:$(VERSION)"  --file=./build/Dockerfile .

docker.darwin:
	docker build  --tag="nsfearthcube/gleaner:$(VERSION)"  --file=./build/Dockerfile .

dockerpush:
	podman push localhost/nsfearthcube/gleaner:$(VERSION) fils/gleaner:$(VERSION)
	podman push localhost/nsfearthcube/gleaner:$(VERSION) fils/gleaner:latest

dockerpush.darwin:
	docker push localhost/nsfearthcube/gleaner:$(VERSION) fils/gleaner:$(VERSION)
	docker push localhost/nsfearthcube/gleaner:$(VERSION) fils/gleaner:latest

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
