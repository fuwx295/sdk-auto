package instrument

import (
	"context"
	"fmt"
	"log"
	"os"
	"sdk-auto/pkg/process"

	auto "go.opentelemetry.io/auto"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

type GoInstrumentorFactory struct{}

func (g *GoInstrumentorFactory) NewInstrumentor(ctx context.Context, endPoint string, details *process.Details) (SdkInstrumentor, error) {
	defaultExporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint(endPoint),
	)
	if err != nil {
		log.Printf("failed to create exporter: %+v", err)
		return nil, err
	}
	inst, err := auto.NewInstrumentation(
		ctx,
		auto.WithPID(details.ProcessID),
		auto.WithServiceName(details.Comm),
		auto.WithTraceExporter(defaultExporter),
		auto.WithGlobal(),
	)
	if err != nil {
		log.Printf("instrumentation setup failed: %v", err)
		return nil, err
	}

	checkRestart(details.ProcessID)
	return &GoInstrumentor{inst: inst}, nil
}

type GoInstrumentor struct {
	inst *auto.Instrumentation
}

func (g *GoInstrumentor) Run(ctx context.Context) error {
	return g.inst.Run(ctx)
}

func (g *GoInstrumentor) Close(ctx context.Context) error {
	return g.inst.Close()
}

func checkRestart(pid int) {
	// 存在探针重启导致 没有将/sys/fs/bpf/{pid}删除，导致再次Instrument，报错file exists
	bpfFile := fmt.Sprintf("/sys/fs/bpf/%d", pid)
	if _, err := os.Stat(bpfFile); os.IsNotExist(err) {
		return
	}
	// 此处如果发现已存在，则考虑先清除目录
	log.Printf("Clean Bpf Folder: %s", bpfFile)
	os.RemoveAll(bpfFile)
}
