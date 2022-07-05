.PHONY: build
build:
	CGO_ENABLED=0 go build -a -installsuffix cgo -o out/gitlab-approval ./main.go

buildlinux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o out/gitlab-approval ./main.go
