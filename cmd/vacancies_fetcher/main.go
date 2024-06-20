package vacancies_fetcher

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"log"
	"sync"
	"time"
	pb "vacancies/protobuf"
	"vacancies/vacancies_fetcher"
	"vacancies/vacancies_fetcher/config"
	"vacancies/vacancies_fetcher/repository"
)

func main() {
	// grpc connection
	cfg := config.MustLoad()
	target := cfg.AuthServiceName + ":" + cfg.AuthServicePort

	conn, err := grpc.NewClient(target)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func() {
		if connErr := conn.Close(); connErr != nil && err == nil {
			log.Fatal(connErr)
		}
	}()
	c := pb.NewAuthServiceClient(conn)

	// obtain cookies
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	authCookies, err := c.RequestAuthorizationData(ctx, &pb.Request{Message: "cookies pls"})
	if err != nil {
		log.Fatalf("could not get authorization data: %v", err)
	}

	// check cookies
	cookies := []*pb.Cookie{
		authCookies.CheckCookies,
		authCookies.Mid,
		authCookies.XCareerSession,
		authCookies.RememberUserToken,
	}

	success, err := vacancies_fetcher.CheckAuth(cookies)
	if err != nil || !success {
		log.Fatalf("could not check auth: %v", err)
	}
	log.Println("Authorization succeeded")
	// fetcher dependency
	fetcher := vacancies_fetcher.New(cfg.KeyWord, authCookies)

	// repo dependency
	pxgConfig, err := pgxpool.ParseConfig(cfg.Dsn)
	if err != nil {
		log.Fatal(err)
	}

	dbPool, err := pgxpool.NewWithConfig(ctx, pxgConfig)
	if err != nil {
		log.Fatal(err)
	}

	db := repository.New(dbPool)

	repo := repository.NewRepo(repository.Deps{
		Repository: db,
		TxBuilder:  dbPool,
	})

	// message broker dependency
	rabbitConn, err := amqp.Dial(cfg.MessageBroker)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConn.Close()

	ch, err := rabbitConn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	messageBroker := vacancies_fetcher.NewRabbit(cfg, ch)

	client := vacancies_fetcher.NewClient(vacancies_fetcher.Deps{
		Fetcher:       fetcher,
		MessageBroker: messageBroker,
		Repo:          repo,
	})
	var wg sync.WaitGroup
	wg.Add(1)
	err = client.Start(cfg, &wg)
	if err != nil {
		log.Fatal(err)
	}
	wg.Wait()
}
