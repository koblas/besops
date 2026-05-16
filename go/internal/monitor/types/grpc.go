package types

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/bufbuild/protocompile"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/koblas/besops/internal/monitor"
	"github.com/koblas/besops/lib/status"
)

type GRPCChecker struct{}

func (c *GRPCChecker) Type() string { return "grpc-keyword" }

func (c *GRPCChecker) Check(ctx context.Context, cfg *monitor.Config) (monitor.CheckResult, error) {
	if cfg.GRPC.URL == "" {
		return monitor.CheckResult{
			Status:  status.Down,
			Message: "gRPC URL is required",
		}, nil
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var dialOpts []grpc.DialOption
	if cfg.GRPC.EnableTLS {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: cfg.IgnoreTLS, //nolint:gosec // user-configurable
		})))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	start := time.Now()
	conn, err := grpc.NewClient(cfg.GRPC.URL, dialOpts...)
	if err != nil {
		return monitor.CheckResult{
			Status:  status.Down,
			Message: fmt.Sprintf("gRPC dial failed: %v", err),
		}, nil
	}
	defer conn.Close()

	methodDesc, err := resolveMethod(ctx, cfg.GRPC.Protobuf, cfg.GRPC.ServiceName, cfg.GRPC.Method)
	if err != nil {
		return monitor.CheckResult{
			Status:  status.Down,
			Message: fmt.Sprintf("protobuf parse error: %v", err),
		}, nil
	}

	inputMsg := dynamicpb.NewMessage(methodDesc.Input())
	if cfg.GRPC.Body != "" {
		if unmarshalErr := protojson.Unmarshal([]byte(cfg.GRPC.Body), inputMsg); unmarshalErr != nil {
			return monitor.CheckResult{
				Status:  status.Down,
				Message: fmt.Sprintf("failed to parse gRPC body: %v", unmarshalErr),
			}, nil
		}
	}

	fullMethod := fmt.Sprintf("/%s/%s", methodDesc.Parent().FullName(), methodDesc.Name())
	outputMsg := dynamicpb.NewMessage(methodDesc.Output())

	invokeErr := conn.Invoke(ctx, fullMethod, inputMsg, outputMsg)
	ping := time.Since(start).Milliseconds()

	var response string
	if invokeErr != nil {
		response = invokeErr.Error()
	} else {
		respBytes, _ := protojson.Marshal(outputMsg)
		response = string(respBytes)
	}

	if cfg.Keyword != "" {
		keywordFound := strings.Contains(response, cfg.Keyword)
		expectFound := cfg.KeywordType != "not contain"

		if keywordFound != expectFound {
			truncated := response
			if len(truncated) > 50 {
				truncated = truncated[:47] + "..."
			}
			return monitor.CheckResult{
				Status:  status.Down,
				Ping:    ping,
				Message: fmt.Sprintf("keyword [%s] %s in response: %s", cfg.Keyword, boolToPresence(!expectFound), truncated),
			}, nil
		}

		return monitor.CheckResult{
			Status:  status.Up,
			Ping:    ping,
			Message: fmt.Sprintf("keyword [%s] %s", cfg.Keyword, boolToPresence(expectFound)),
		}, nil
	}

	if invokeErr != nil {
		return monitor.CheckResult{
			Status:  status.Down,
			Ping:    ping,
			Message: fmt.Sprintf("gRPC call failed: %v", invokeErr),
		}, nil
	}

	return monitor.CheckResult{
		Status:  status.Up,
		Ping:    ping,
		Message: response,
	}, nil
}

func boolToPresence(found bool) string {
	if found {
		return "found"
	}
	return "not found"
}

func resolveMethod(ctx context.Context, protoSource, serviceName, methodName string) (protoreflect.MethodDescriptor, error) {
	if protoSource == "" {
		return nil, fmt.Errorf("protobuf definition is required")
	}

	compiler := &protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(
			&protocompile.SourceResolver{
				Accessor: protocompile.SourceAccessorFromMap(map[string]string{
					"input.proto": protoSource,
				}),
			},
		),
	}

	files, err := compiler.Compile(ctx, "input.proto")
	if err != nil {
		return nil, fmt.Errorf("compiling proto: %w", err)
	}

	fd := files[0]
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		svc := services.Get(i)
		if string(svc.FullName()) == serviceName || string(svc.Name()) == serviceName {
			methods := svc.Methods()
			for j := 0; j < methods.Len(); j++ {
				m := methods.Get(j)
				if string(m.Name()) == methodName {
					return m, nil
				}
			}
			return nil, fmt.Errorf("method %s not found in service %s", methodName, serviceName)
		}
	}

	return nil, fmt.Errorf("service %s not found in proto", serviceName)
}
