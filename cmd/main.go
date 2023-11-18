package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"stock/ohlc/handler"
	"stock/ohlc/repository"
	"stock/ohlc/service"
	"stock/pb"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func grpcServer(ctx context.Context, ohlcGrpcHandler pb.OHLCServer) error {
	lis, err := net.Listen("tcp", os.Getenv("GRPC_ADDRESS"))
	if err != nil {
		log.Println("errors.failed to listen TCP for gRPC connection", err)
		return err
	}

	grpcServer := grpc.NewServer()
	// register servers
	pb.RegisterOHLCServer(grpcServer, ohlcGrpcHandler)

	log.Println("info.serving gRPC on connection ")
	go func() {
		log.Println(grpcServer.Serve(lis))
	}()

	conn, err := grpc.Dial(":9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln("errors.failed to dial GRPC server", err)
		return err
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Println("errors.closing connection gRPC dial", err)
		}
	}()

	mux := runtime.NewServeMux()
	// register handlers
	err = pb.RegisterOHLCHandler(ctx, mux, conn)
	if err != nil {
		log.Println("errors.failed to register gateway")
		return err
	}

	gwServer := &http.Server{
		ReadHeaderTimeout: time.Second * 10,
		Addr:              os.Getenv("HTTP_ADDRESS"),
		Handler:           mux,
	}

	log.Println("info.serving gRPC-Gateway on connection")
	log.Println(gwServer.ListenAndServe())

	return nil
}

func createKafkaProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{os.Getenv("REDPANDA_ADDRESS")}, config)
	if err != nil {
		log.Println("errors.sync producer connection")
		return nil, err
	}

	return producer, nil
}

var wg sync.WaitGroup

func kafkaConsumer(topic string, cb func(key, value []byte) error) error {
	consumer, err := sarama.NewConsumer([]string{os.Getenv("REDPANDA_ADDRESS")}, nil)
	if err != nil {
		log.Println("errors.failed to start consumer")
		return err
	}
	// get all paritions idx
	partitionList, err := consumer.Partitions(topic)
	if err != nil {
		log.Println("errors.Failed to get the list of partitions")
		return err
	}

	// listen to all parititions
	for partition := range partitionList {
		cp, err := consumer.ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			log.Printf("errors.failed to start consumer for partition %d: %s\n", partition, err)
			return err
		}
		defer cp.AsyncClose()
		wg.Add(1)
		go func(pc sarama.PartitionConsumer) {
			defer wg.Done()
			for msg := range pc.Messages() {
				err := cb(msg.Key, msg.Value)
				if err != nil {
					log.Println("errors.when processing msg")
				}
			}
		}(cp)
	}

	wg.Wait()
	return consumer.Close()
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	kafkaProducer, err := createKafkaProducer()
	if err != nil {
		log.Panicln(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("REDIS_AUTH"),
	})

	ohlcRepository := repository.NewOHLCRepository(rdb)
	ohlcService := service.NewOHLCService(ohlcRepository, kafkaProducer)
	ohlcGrpcHandler := handler.NewOHLCgRPCHandler(ohlcService)

	go func() {
		log.Println("info.start kafka consumer")
		err := kafkaConsumer("ohlc", ohlcService.ListenOrderTx)
		if err != nil {
			log.Println("errors.starting consumer ")
		}
	}()

	log.Println("info.start grpc server..")
	if err := grpcServer(ctx, ohlcGrpcHandler); err != nil {
		log.Panicln(err)
	}
}
