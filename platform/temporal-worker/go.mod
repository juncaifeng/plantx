module github.com/plantx/platform/temporal-worker

go 1.24.0

toolchain go1.24.2

require (
	github.com/plantx/platform/registry-service/api v0.0.0-00010101000000-000000000000
	go.temporal.io/sdk v1.26.0
	google.golang.org/grpc v1.70.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/mock v1.7.0-rc.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/plantx/kit/kit-go v0.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	go.temporal.io/api v1.29.1 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/time v0.10.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/plantx/kit/kit-go => ../../kit/kit-go
	github.com/plantx/kit/kit-go/gateway => ../../kit/kit-go/gateway
	github.com/plantx/platform/registry-service/api => ../registry-service/api
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20250303144028-a0af3efb3deb
)
