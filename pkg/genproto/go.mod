module github.com/wosai/ultron/v2/pkg/genproto

go 1.17

require (
	github.com/golang/protobuf v1.5.0 // indirect
	github.com/wosai/ultron/v2/pkg/statistics v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd // indirect
	golang.org/x/text v0.3.0 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)

replace github.com/wosai/ultron/v2/pkg/statistics => ../statistics
