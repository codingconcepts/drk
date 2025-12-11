test:
	go test ./... -v -cover

cover:
	go test ./... -coverprofile=cover.out
	go tool cover -html=cover.out

validate_version:
ifndef VERSION
	$(error VERSION is undefined)
endif

docker_push: validate_version
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" drk.go

	@docker build --platform linux/amd64 \
		-t codingconcepts/drk:linux_amd64_${VERSION} \
		--push \
		.

	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=${VERSION}" drk.go

	@docker build --platform linux/arm64 \
		-t codingconcepts/drk:linux_arm64_${VERSION} \
		--push \
		.
	
	rm ./drk

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

	open releases

teardown:
	docker ps -aq | xargs docker rm -f