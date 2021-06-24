version="0.1.2"

build-go-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags="-X main.Version=${version}" -o package/build/tprelay cmd/server/main.go

package-docker:build-go-amd64
	docker build -t thingsplex/tprelay:${version} -t thingsplex/tprelay:latest .

publish-docker:package-docker
	docker push thingsplex/tprelay:${version}
	docker push thingsplex/tprelay:latest

docker-run:
	docker run -p 80:80 --name tprelay  thingsplex/tprelay:latest

gen-proto:
	 protoc -I=./protodef --go_out=./pkg/proto ./protodef/tunnel_frame.proto