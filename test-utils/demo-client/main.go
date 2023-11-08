package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func main() {
	// Attempt to load .env
	_ = godotenv.Load(".env")

	cfg, err := ParseArgs()
	if err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}

	if cfg.Debug {
		logrus.Info("debug mode enabled")
		logrus.SetLevel(logrus.DebugLevel)
	}

	displayConfig(cfg)

	// Start DataDog tracer
	if os.Getenv("DD_ENV") != "" {
		tracer.Start(tracer.WithAnalytics(true))
		defer tracer.Stop()

		err := profiler.Start(
			profiler.WithService("demo-service"),
			profiler.WithEnv(os.Getenv("DD_ENV")),
			profiler.WithProfileTypes(
				profiler.CPUProfile,
				profiler.HeapProfile,
			),
		)
		if err != nil {
			log.Fatal(err)
		}

		defer profiler.Stop()

	}

	go func() {
		http.ListenAndServe("localhost:8888", nil)
	}()

	if err := Run(cfg); err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}
}

func displayConfig(cfg *Config) {
	if cfg == nil {
		return
	}

	logrus.Info("demo client settings:")
	logrus.Infof("  server address: %s", cfg.ServerAddress)
	logrus.Infof("  server token: %s", cfg.ServerToken)
	logrus.Infof("  service name: %s", cfg.ServiceName)
	logrus.Infof("  operation type: %d", cfg.OperationType)
	logrus.Infof("  operation name: %s", cfg.OperationName)
	logrus.Infof("  component name: %s", cfg.ComponentName)
	logrus.Infof("  num instances: %d", cfg.NumInstances)
	logrus.Infof("  reconnect random: %t", cfg.ReconnectRandom)
	logrus.Infof("  reconnect interval: %d", cfg.ReconnectInterval)
	logrus.Infof("  message rate: %v", cfg.MessageRate)
	logrus.Infof("  data source type: %s", cfg.DataSourceType)
	logrus.Infof("  data source file: %s", cfg.DataSourceFile.Name())
	logrus.Infof("  debug: %t", cfg.Debug)
	logrus.Infof("  quiet: %t", cfg.Quiet)
	logrus.Infof("  inject logger: %t", cfg.InjectLogger)
}
