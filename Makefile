
gen-proto:
	 protoc -I=./protodef --go_out=./pkg/proto ./protodef/tunnel_frame.proto