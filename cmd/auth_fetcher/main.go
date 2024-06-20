package auth_fetcher

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	"vacancies/auth_fetcher"
	"vacancies/auth_fetcher/config"
	pb "vacancies/protobuf"
)

func main() {
	cfg := config.MustLoad()

	// login
	auth := auth_fetcher.NewAuth(cfg)
	authCookies, err := auth.Login()
	if err != nil {
		log.Fatalf("could not login: %v", err)
	}

	// server
	srv := server{
		CookiesToSend: authCookies,
	}
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, srv)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type server struct {
	pb.UnimplementedAuthServiceServer
	CookiesToSend map[string]*pb.Cookie
}

func (s server) RequestAuthorizationData(ctx context.Context, request *pb.Request) (*pb.AuthorizationData, error) {
	log.Println("Sending authorization data")
	return &pb.AuthorizationData{
		XCareerSession:    s.CookiesToSend["_career_session"],
		RememberUserToken: s.CookiesToSend["remember_user_token"],
		CheckCookies:      s.CookiesToSend["check_cookies"],
		Mid:               s.CookiesToSend["mid"],
	}, nil
}
