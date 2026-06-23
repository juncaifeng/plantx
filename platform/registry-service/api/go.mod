module github.com/plantx/platform/registry-service/api

go 1.24.0

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3
	github.com/plantx/kit/kit-go v0.0.0
	google.golang.org/genproto/googleapis/api v0.0.0-20250303144028-a0af3efb3deb
	google.golang.org/grpc v1.70.0
	google.golang.org/protobuf v1.36.5
)

require (
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250303144028-a0af3efb3deb // indirect
)

replace github.com/plantx/kit/kit-go => ../../../kit/kit-go
