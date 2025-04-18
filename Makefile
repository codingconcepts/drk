test:
	go test ./... -v -cover

cover:
	go test ./... -coverprofile=cover.out
	go tool cover -html=cover.out

validate_version:
ifndef VERSION
	$(error VERSION is undefined)
endif

docker_build: validate_version
	docker build \
		-t codingconcepts/drk:${VERSION} \
		--build-arg version=${VERSION} \
		.

docker_push: docker_build
	docker push codingconcepts/drk:${VERSION}
	docker tag codingconcepts/drk:${VERSION} codingconcepts/drk:latest
	docker push codingconcepts/drk:latest

release: validate_version
	- mkdir releases

	# linux (amd)
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" -o drk ;\
	tar -zcvf ./releases/drk_${VERSION}_linux_amd64.tar.gz ./drk ;\

	# macos (arm)
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=${VERSION}" -o drk ;\
	tar -zcvf ./releases/drk_${VERSION}_macos_arm64.tar.gz ./drk ;\

	# macos (amd)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" -o drk ;\
	tar -zcvf ./releases/drk_${VERSION}_macos_amd64.tar.gz ./drk ;\

	# windows (amd)
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" -o drk ;\
	tar -zcvf ./releases/drk_${VERSION}_windows_amd64.tar.gz ./drk ;\

	rm ./drk

teardown:
	docker ps -aq | xargs docker rm -f