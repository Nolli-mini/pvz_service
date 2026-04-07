package grpc

import (
	"context"
	"log"
	"net"
	pvz_v1 "pvz-service/internal/grpc/pvz"
	"pvz-service/internal/repository"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PVZServer struct {
	pvz_v1.UnimplementedPVZServiceServer
	repo *repository.Repository
}

func NewPVZServer(repo *repository.Repository) *PVZServer {
	return &PVZServer{
		repo: repo,
	}
}

func (s *PVZServer) GetPVZList(ctx context.Context, req *pvz_v1.GetPVZListRequest) (*pvz_v1.GetPVZListResponse, error) {

	pvzs, _, err := s.repo.GetPVZList(ctx, nil, nil, 1, 1000)
	if err != nil {
		return nil, err
	}

	response := &pvz_v1.GetPVZListResponse{
		Pvzs: make([]*pvz_v1.PVZ, 0, len(pvzs)),
	}

	for _, pvz := range pvzs {
		response.Pvzs = append(response.Pvzs, &pvz_v1.PVZ{
			Id:               pvz.Id.String(),
			RegistrationDate: timestamppb.New(*pvz.RegistrationDate),
			City:             string(pvz.City),
		})
	}

	return response, nil
}

func StartGRPCServer(repo *repository.Repository) {
	lis, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatalf("Failed to listen on port 3000: %v", err)
	}

	grpcServer := grpc.NewServer()
	pvz_v1.RegisterPVZServiceServer(grpcServer, NewPVZServer(repo))

	reflection.Register(grpcServer)

	log.Println("gRPC server starting on :3000")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
