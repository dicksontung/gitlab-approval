all: build build-linux compress

.PHONY: build
build:
	go build -a -installsuffix cgo -o out/gitlab-approval ./main.go

.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o out/gitlab-approval-linux-amd64 ./main.go

.PHONY: build-docker
build-docker:
	docker build -t dixont/gitlab-approval

compress:
	tar -czvf out/gitlab-approval-linux-amd64.tar.gz --directory=out/ gitlab-approval-linux-amd64