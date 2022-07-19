FROM golang:alpine as builder
RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY out/gitlab-approval-linux-amd64 /app/gitlab-approval
WORKDIR app
EXPOSE 8080

CMD ["./gitlab-approval", "webhook", "--port", "8080"]