module github.com/wosai/ultron/v2

require (
	github.com/go-chi/chi/v5 v5.0.4
	github.com/golang/protobuf v1.5.0
	github.com/google/uuid v1.3.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/stretchr/testify v1.7.0
	github.com/wosai/ultron/v2/pkg/genproto v0.0.0-00010101000000-000000000000
	github.com/wosai/ultron/v2/pkg/statistics v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.19.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/net v0.0.0-20210510120150-4163338589ed // indirect
	golang.org/x/sys v0.0.0-20210514084401-e8d321eab015 // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace (
	github.com/wosai/ultron/v2/pkg/genproto => ./pkg/genproto
	github.com/wosai/ultron/v2/pkg/statistics => ./pkg/statistics
)

go 1.17
