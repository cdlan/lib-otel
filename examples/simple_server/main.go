package main

import (
	"context"
	"log"
	"net"

	keycloak "github.com/cdlan/lib-keycloak"
	
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"github.com/spf13/viper"
	"google.golang.org/grpc/reflection"
)

var Conf keycloak.Config

const (
	serviceName string = "my-service"
	version     string = "v1.0.0"
)

func init() {

	// set to default
	Conf.Default()

	// load from yaml
	viper.SetConfigName("config.yml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")

	if err := viper.ReadInConfig(); err != nil {
		log.Println("Config file not found!")
	}

	if err := viper.Unmarshal(&Conf); err != nil {
		log.Println("Viper failed to unmarshall config: ", err)
	}

	// load from env
	Conf.LoadVarsFromEnv()

	// init otel
	err := Conf.Init(serviceName, version)
	if err != nil {
		panic(err)
	}

}

func main() {

	// close OTEL
	defer func() {
		if err := Conf.ShutdownTracerProvider(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	/* setup grpc */

	lis, err := net.Listen("tcp", ":4445")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	// register servers
	healthgrpc.RegisterHealthServer(grpcServer, health.NewServer())

	// add reflection
	reflection.Register(grpcServer)

	// start listening
	if err := grpcServer.Serve(lis); err != nil {

		log.Fatalf("Failed to serve grpc: %v", err)
	}
}