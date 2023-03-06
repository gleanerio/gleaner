BINARY := gleaner
BINARYIO := glcon
VERSION :=`cat VERSION`
MAINVERSION :=`cat ../../VERSION`
.DEFAULT_GOAL := gleaner

gleaner:
	cd cmd/$(BINARY) ; \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 env go build -ldflags "-X main.VERSION=$(MAINVERSION)" -o $(BINARY);\
    cp $(BINARY) ../../


gleaner.exe:
	cd cmd/$(BINARY) ; \
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 env go build -ldflags "-X main.VERSION=$(MAINVERSION)" -o $(BINARY).exe;\
    cp $(BINARY).exe ../../

gleaner.darwin:
	cd cmd/$(BINARY) ; \
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 env go build  -ldflags "-X main.VERSION=$(MAINVERSION)" -o $(BINARY)_darwin;\
    cp $(BINARY)_darwin ../../

gleaner.m2:
	cd cmd/$(BINARY) ; \
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 env go build -ldflags "-X main.VERSION=$(MAINVERSION)" -o $(BINARY)_m2;\
    cp $(BINARY)_m2 ../../

glcon:
	cd cmd/$(BINARYIO) ; \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 env go build -ldflags "-X main.VERSION=$(MAINVERSION)" -o $(BINARYIO);\
    cp $(BINARYIO) ../../

glcon.exe:
	cd cmd/$(BINARYIO) ; \
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 env go build -ldflags "-X main.VERSION=$(MAINVERSION)" -o $(BINARYIO).exe;\
    cp $(BINARYIO).exe ../../

glcon.darwin:
	cd cmd/$(BINARYIO) ; \
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 env go build -ldflags "-X main.VERSION=$(MAINVERSION)" -o $(BINARYIO)_darwin ;\
	cp $(BINARYIO)_darwin ../../

glcon.m2:
	cd cmd/$(BINARYIO) ; \
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 env go build -ldflags "-X main.VERSION=$(MAINVERSION)" -o $(BINARYIO)_m2;\
	cp $(BINARYIO)_m2 ../../

releases: gleaner gleaner.exe gleaner.darwin gleaner.m2 glcon glcon.exe glcon.darwin glcon.m2

docker:
	podman build  --tag="nsfearthcube/gleaner:$(VERSION)"  --file=./build/Dockerfile .

docker.multiarch: gleaner
	docker buildx build --no-cache --pull --platform=linux/arm64,linux/amd64 --push -t nsfearthcube/gleaner:$(VERSION)  --file=./build/Dockerfile .

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
