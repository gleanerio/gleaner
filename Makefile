BINARY := gleaner
VERSION :=`cat VERSION`
.DEFAULT_GOAL := gleaner

gleaner:
	cd cmd/$(BINARY) ; \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 env go build -o $(BINARY)

glcon:
	cd cmd/glcon ; \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 env go build -o glcon

docker:
	docker build  --tag="earthcube/gleaner:$(VERSION)"  --file=./build/Dockerfile . ; \
	docker tag earthcube/gleaner:$(VERSION) earthcube/gleaner:latest

removeimage:
	docker rmi --force earthcube/gleaner:$(VERSION)
	docker rmi --force earthcube/gleaner:latest

publish: docker
	docker push earthcube/gleaner:$(VERSION) ; \
	docker push earthcube/gleaner:latest
