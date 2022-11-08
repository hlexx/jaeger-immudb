package trace

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"testing"
)

const maxValSize = 1 << 20

func TestTrace(t *testing.T) {
	t.Setenv("JAEGER_URL", "http://localhost:24268/api/traces")
	type service struct {
		name string
		args map[string]string
		err  error
	}
	type args struct {
		id       string
		services []service
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "add trace to jaeger",
			args: args{
				id: uuid.NewString(),
				services: []service{
					{name: "test", args: map[string]string{}},
					{name: "dev", args: map[string]string{}},
					{name: "prod", args: map[string]string{}},
					{name: "prod", args: map[string]string{}},
					{name: "database", args: map[string]string{}, err: errors.New("not connected")},
					{name: "svc1", args: map[string]string{}},
					{name: "svc2", args: map[string]string{}},
					{name: "svc3", args: map[string]string{}},
					{name: "svc4", args: map[string]string{}},
					{name: "svc5", args: map[string]string{}},
					{name: "svc6", args: map[string]string{}},
					{name: "svc7", args: map[string]string{}},
					{name: "svc8", args: map[string]string{}},
					{name: "svc9", args: map[string]string{}},
					{name: "svc10", args: map[string]string{}},
					{name: "svc11", args: map[string]string{}},
					{name: "svc12", args: map[string]string{}},
					{name: "svc13", args: map[string]string{}},
					{name: "svc14", args: map[string]string{}},
					{name: "svc15", args: map[string]string{}},
					{name: "svc16", args: map[string]string{}},
					{name: "svc17", args: map[string]string{}},
					{name: "svc18", args: map[string]string{}},
					{name: "svc19", args: map[string]string{}},
					{name: "svc20", args: map[string]string{}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, service := range tt.args.services {
				service.args["inc"] = fmt.Sprintf("%v", i)
				if service.err != nil {
					service.args[ErrorKey] = service.err.Error()
				}
				Trace(tt.args.id, service.name, service.args)
			}
		})
	}
}
