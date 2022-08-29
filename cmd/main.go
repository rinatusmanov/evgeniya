package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
	jaegerZap "github.com/uber/jaeger-client-go/log/zap"
	jProm "github.com/uber/jaeger-lib/metrics/prometheus"
	"gitlab.com/pjrpc/pjrpc/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapio"

	_ "github.com/lib/pq"

	"github.com/rinatusmanov/jsonrpc20/internal/pkg/seamlessv2"
	"github.com/rinatusmanov/jsonrpc20/internal/pkg/seamlessv2/generated"
)

func main() {
	// инициализация логгера
	atom := zap.NewAtomicLevel()
	logger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.Lock(os.Stdout),
			atom,
		),
	).With(zap.String("service", "SeamlessV2ServiceServer"))

	atom.SetLevel(zap.InfoLevel)

	// инициализация клиента jaeger
	cfg, errFromEnv := jaegerCfg.FromEnv()
	if errFromEnv != nil {
		logger.Panic("Could not parse Jaeger env vars", zap.Error(errFromEnv))
	}

	cfg.ServiceName = "SeamlessV2ServiceServer"

	factory := jProm.New(jProm.WithRegisterer(prometheus.NewPedanticRegistry()))

	tracer, closer, errNewTracer := cfg.NewTracer(jaegerCfg.Metrics(factory))
	if errFromEnv != nil {
		logger.Panic("Could not parse Jaeger env vars", zap.Error(errNewTracer))
	}

	defer func() {
		_ = closer.Close()
	}()

	opentracing.SetGlobalTracer(jaegerZap.NewLoggingTracer(logger, tracer))

	srv := pjrpc.NewServerHTTP()
	srv.SetLogger(&zapio.Writer{Log: logger, Level: zapcore.InfoLevel})

	db, errOpen := sqlx.Open("postgres", os.Getenv("CONNECTION_STRING"))
	if errOpen != nil {
		logger.Panic("Could not open database", zap.Error(errOpen))
	}

	rpcService := seamlessv2.NewRPCService(db, logger)

	generated.RegisterSeamlessV2ServiceServer(srv, rpcService, TraceMiddleWare)

	http.Handle("/rpc/", srv)

	if errListenAndServe := http.ListenAndServe(":8086", nil); errListenAndServe != nil {
		panic(errListenAndServe)
	}
}

func TraceMiddleWare(next pjrpc.Handler) pjrpc.Handler {
	return func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		tracer := opentracing.GlobalTracer()

		span := tracer.StartSpan("RPCService.MiddleWare")

		defer span.Finish()

		now := time.Now()
		res, err := next(opentracing.ContextWithSpan(ctx, span), params)
		dur := time.Since(now).Seconds()

		span.SetTag("duration", dur)
		span.SetTag("error", err)
		span.SetTag("result", res)

		return res, err
	}
}
