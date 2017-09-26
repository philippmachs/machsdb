GOARCH = amd64
GOOS = linux
CGO_ENABLED = 0

all:
	GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=${CGO_ENABLED} go build -a -installsuffix cgo -o cache examples/server/server.go ;\
	chmod +x cache ;\
	docker build -f Dockerfile -t kv_cache . ;\
	rm cache
