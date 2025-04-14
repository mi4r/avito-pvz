package grpc

import (
	"context"
	"net"
	"time"

	pvz_v1 "github.com/mi4r/avito-pvz/api/pvz/v1"
	"github.com/mi4r/avito-pvz/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pvz_v1.UnimplementedPVZServiceServer
	store storage.Storage
}

func NewServer(store storage.Storage) *Server {
	return &Server{
		store: store,
	}
}

func (s *Server) Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pvz_v1.RegisterPVZServiceServer(grpcServer, s)

	return grpcServer.Serve(lis)
}

func (s *Server) GetPVZList(ctx context.Context, req *pvz_v1.GetPVZListRequest) (*pvz_v1.GetPVZListResponse, error) {
	// Get all PVZs from storage
	pvzs, err := s.store.GetPVZsWithReceptions(ctx, time.Now().Add(-24*time.Hour), time.Now(), 1, 1000)
	if err != nil {
		return nil, err
	}

	// Convert storage PVZs to gRPC PVZs
	grpcPVZs := make([]*pvz_v1.PVZ, 0, len(pvzs))
	for _, pvz := range pvzs {
		grpcPVZs = append(grpcPVZs, &pvz_v1.PVZ{
			Id:               pvz.PVZ.ID.String(),
			RegistrationDate: timestamppb.New(pvz.PVZ.RegistrationDate),
			City:             pvz.PVZ.City,
		})
	}

	return &pvz_v1.GetPVZListResponse{
		Pvzs: grpcPVZs,
	}, nil
}
