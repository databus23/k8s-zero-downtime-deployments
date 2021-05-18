FROM golang:alpine
WORKDIR /app
ADD . /app
RUN --mount=type=cache,target=/go/pkg/mod \
	  --mount=type=cache,target=/root/.cache/go-build \
    go build -o /test-server

FROM alpine
LABEL source_repository=github.com/databus23/k8s-zero-downtime-deployments
COPY --from=0 /test-server /test-server
ENTRYPOINT ["/test-server"]
