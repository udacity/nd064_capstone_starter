module github.com/GoogleCloudPlatform/microservices-demo/src/productcatalogservice

go 1.15

require (
	cloud.google.com/go/profiler v0.3.1
	contrib.go.opencensus.io/exporter/jaeger v0.2.0
	contrib.go.opencensus.io/exporter/stackdriver v0.5.0
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.9
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/uber/jaeger-client-go v2.21.1+incompatible // indirect
	go.opencensus.io v0.24.0
	golang.org/x/net v0.5.0
	google.golang.org/grpc v1.53.0
)
