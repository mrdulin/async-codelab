package main

import (
	"fmt"
	"log"
	"net"

	"github.com/mrdulin/grpc-go-cnode/configs"
	"github.com/mrdulin/grpc-go-cnode/internal/protobufs/topic"
	"github.com/mrdulin/grpc-go-cnode/internal/protobufs/user"
	"github.com/mrdulin/grpc-go-cnode/internal/utils/grpclogger"
	"github.com/mrdulin/grpc-go-cnode/internal/utils/http"
	"github.com/mrdulin/grpc-go-cnode/internal/utils/interceptors"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	conf   *viper.Viper
	logger grpclog.LoggerV2
)

func init() {
	conf = configs.Read()
	logger = grpclogger.New(conf.GetString(configs.GRPC_GO_LOG_SEVERITY_LEVEL), conf.GetString(configs.GRPC_GO_LOG_VERBOSITY_LEVEL))
}

func main() {
	port := conf.GetString(configs.PORT)
	baseurl := conf.GetString(configs.BASE_URL)
	if baseurl == "" {
		log.Fatal("missing api url")
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	creds, err := credentials.NewServerTLSFromFile("./assets/server.crt", "./assets/server.key")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(interceptors.NewUnaryInterceptor(logger)),
	)
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, hs)

	httpClient := http.NewClient()
	userServiceImpl := user.NewUserServiceImpl(httpClient, baseurl)
	topicServiceImpl := topic.NewTopicServiceImpl(httpClient, baseurl)
	user.RegisterUserServiceServer(grpcServer, userServiceImpl)
	topic.RegisterTopicServiceServer(grpcServer, topicServiceImpl)

	log.Printf("gRPC server is listening on port: %s\n", port)
	log.Fatal(grpcServer.Serve(lis))
}
