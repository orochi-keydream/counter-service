package app

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/orochi-keydream/counter-service/internal/api"
	"github.com/orochi-keydream/counter-service/internal/config"
	"github.com/orochi-keydream/counter-service/internal/jobs"
	"github.com/orochi-keydream/counter-service/internal/kafka/consumer"
	"github.com/orochi-keydream/counter-service/internal/kafka/producer"
	"github.com/orochi-keydream/counter-service/internal/proto/admin"
	"github.com/orochi-keydream/counter-service/internal/proto/counter"
	"github.com/orochi-keydream/counter-service/internal/repository"
	"github.com/orochi-keydream/counter-service/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func Run() {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := config.LoadConfig()

	realtimeConfigService := service.NewRealtimeConfigService()

	conn := newConn(cfg.Database)

	messageRepository := repository.NewMessageRepository(conn)
	commandRepository := repository.NewCommandRepository(conn)
	outboxRepository := repository.NewOutboxRepository(conn)
	transactionManager := repository.NewTransactionManager(conn)

	commandProducer, err := producer.NewDialogueCommandProducer(cfg.Kafka)

	counterService := service.NewCounterService(
		messageRepository,
		commandRepository,
		outboxRepository,
		transactionManager,
		realtimeConfigService,
	)

	outboxService := service.NewOutboxService(outboxRepository, commandProducer)

	wg := &sync.WaitGroup{}

	job := jobs.NewOutboxJob(outboxService)

	wg.Add(1)
	job.Start(ctx, wg)

	commandsConsumer := consumer.NewCounterCommandsConsumer(counterService)

	wg.Add(1)

	err = consumer.RunCounterCommandsConsumer(
		ctx,
		cfg.Kafka.Brokers,
		cfg.Kafka.Consumers.CounterCommands.Topic,
		commandsConsumer,
		wg,
	)

	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	counterGrpcService := api.NewCounterGrpcService(counterService)
	counter.RegisterCounterServiceServer(grpcServer, counterGrpcService)

	adminGrpcService := api.NewAdminGrpcService(realtimeConfigService)
	admin.RegisterCounterServiceServer(grpcServer, adminGrpcService)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Service.GrpcPort))

	if err != nil {
		log.Fatal(err)
	}

	err = grpcServer.Serve(listener)

	log.Println("Serving")

	if err != nil {
		log.Fatal(err)
	}

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)

	select {
	case <-sigterm:
		log.Println("Gracefully shutting down...")
		grpcServer.GracefulStop()
		cancel()
	}

	wg.Wait()

	log.Println("Gracefully shut down")
}

func newConn(cfg config.DatabaseConfig) *sql.DB {
	connStr := fmt.Sprintf(
		"host=%v port=%v user=%v password=%v dbname=%v",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DatabaseName)

	conn, err := sql.Open("pgx", connStr)

	if err != nil {
		panic(err)
	}

	err = conn.Ping()

	if err != nil {
		panic(err)
	}

	return conn
}
