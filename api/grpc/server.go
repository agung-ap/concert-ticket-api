package grpc

import (
	"concert-ticket-api/internal/model"
	"context"
	"fmt"
	"net"
	_ "time"

	"concert-ticket-api/internal/service"
	"concert-ticket-api/pkg/logger"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "concert-ticket-api/api/grpc/proto"
)

// Server represents a gRPC server
type Server struct {
	concertService service.ConcertService
	bookingService service.BookingService
	logger         logger.Logger
	server         *grpc.Server
	port           int
	pb.UnimplementedConcertServiceServer
	pb.UnimplementedBookingServiceServer
}

// NewServer creates a new gRPC server
func NewServer(
	concertService service.ConcertService,
	bookingService service.BookingService,
	logger logger.Logger,
	port int,
) *Server {
	// Create gRPC server with middleware
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_validator.UnaryServerInterceptor(),
		)),
	)

	// Create server instance
	server := &Server{
		concertService: concertService,
		bookingService: bookingService,
		logger:         logger,
		server:         grpcServer,
		port:           port,
	}

	// Register services
	pb.RegisterConcertServiceServer(grpcServer, server)
	pb.RegisterBookingServiceServer(grpcServer, server)

	// Register reflection service (helpful for grpcurl and other tools)
	reflection.Register(grpcServer)

	return server
}

// Start starts the gRPC server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Error("Failed to listen: %v", err)
		return err
	}

	s.logger.Info("Starting gRPC server on %s", addr)
	return s.server.Serve(lis)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	s.logger.Info("Shutting down gRPC server")
	s.server.GracefulStop()
}

// GetConcert implements the ConcertService.GetConcert RPC
func (s *Server) GetConcert(ctx context.Context, req *pb.GetConcertRequest) (*pb.Concert, error) {
	concert, err := s.concertService.GetByID(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get concert: %v", err)
		return nil, err
	}

	return convertModelToPbConcert(concert), nil
}

// ListConcerts implements the ConcertService.ListConcerts RPC
func (s *Server) ListConcerts(ctx context.Context, req *pb.ListConcertsRequest) (*pb.ListConcertsResponse, error) {
	// Convert request to filters
	filters := make(map[string]interface{})

	if req.Artist != "" {
		filters["artist"] = req.Artist
	}

	if req.Venue != "" {
		filters["venue"] = req.Venue
	}

	if req.Name != "" {
		filters["name"] = req.Name
	}

	if req.DateFrom != nil {
		filters["date_from"] = req.DateFrom.AsTime()
	}

	if req.DateTo != nil {
		filters["date_to"] = req.DateTo.AsTime()
	}

	if req.AvailableOnly {
		filters["available"] = true
	}

	// Get concerts
	concerts, totalCount, err := s.concertService.ListConcerts(ctx, int(req.Page), int(req.PageSize), filters)
	if err != nil {
		s.logger.Error("Failed to list concerts: %v", err)
		return nil, err
	}

	// Convert to response
	var pbConcerts []*pb.Concert
	for _, concert := range concerts {
		pbConcerts = append(pbConcerts, convertModelToPbConcert(concert))
	}

	totalPages := (totalCount + int(req.PageSize) - 1) / int(req.PageSize)

	return &pb.ListConcertsResponse{
		Concerts: pbConcerts,
		Meta: &pb.PaginationMeta{
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalCount: int32(totalCount),
			TotalPages: int32(totalPages),
		},
	}, nil
}

// CreateConcert implements the ConcertService.CreateConcert RPC
func (s *Server) CreateConcert(ctx context.Context, req *pb.CreateConcertRequest) (*pb.Concert, error) {
	// Convert request to model
	concert := &model.Concert{
		Name:             req.Name,
		Artist:           req.Artist,
		Venue:            req.Venue,
		ConcertDate:      req.ConcertDate.AsTime(),
		TotalTickets:     int(req.TotalTickets),
		AvailableTickets: int(req.TotalTickets), // Initially, all tickets are available
		Price:            req.Price,
		BookingStartTime: req.BookingStartTime.AsTime(),
		BookingEndTime:   req.BookingEndTime.AsTime(),
	}

	// Create concert
	createdConcert, err := s.concertService.CreateConcert(ctx, concert)
	if err != nil {
		s.logger.Error("Failed to create concert: %v", err)
		return nil, err
	}

	return convertModelToPbConcert(createdConcert), nil
}

// UpdateConcert implements the ConcertService.UpdateConcert RPC
func (s *Server) UpdateConcert(ctx context.Context, req *pb.UpdateConcertRequest) (*pb.Concert, error) {
	// Convert request to model
	concert := &model.Concert{
		ID:               req.Id,
		Name:             req.Name,
		Artist:           req.Artist,
		Venue:            req.Venue,
		ConcertDate:      req.ConcertDate.AsTime(),
		TotalTickets:     int(req.TotalTickets),
		Price:            req.Price,
		BookingStartTime: req.BookingStartTime.AsTime(),
		BookingEndTime:   req.BookingEndTime.AsTime(),
		Version:          int(req.Version),
	}

	// Get current concert to preserve available tickets
	currentConcert, err := s.concertService.GetByID(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get concert for update: %v", err)
		return nil, err
	}

	// Preserve available tickets
	concert.AvailableTickets = currentConcert.AvailableTickets

	// Update concert
	err = s.concertService.UpdateConcert(ctx, concert)
	if err != nil {
		s.logger.Error("Failed to update concert: %v", err)
		return nil, err
	}

	// Get updated concert
	updatedConcert, err := s.concertService.GetByID(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get updated concert: %v", err)
		return nil, err
	}

	return convertModelToPbConcert(updatedConcert), nil
}

// GetBooking implements the BookingService.GetBooking RPC
func (s *Server) GetBooking(ctx context.Context, req *pb.GetBookingRequest) (*pb.Booking, error) {
	booking, err := s.bookingService.GetBookingByID(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get booking: %v", err)
		return nil, err
	}

	return convertModelToPbBooking(booking), nil
}

// GetUserBookings implements the BookingService.GetUserBookings RPC
func (s *Server) GetUserBookings(ctx context.Context, req *pb.GetUserBookingsRequest) (*pb.GetUserBookingsResponse, error) {
	bookings, err := s.bookingService.GetUserBookings(ctx, req.UserId, int(req.Page), int(req.PageSize))
	if err != nil {
		s.logger.Error("Failed to get user bookings: %v", err)
		return nil, err
	}

	// Convert to response
	var pbBookings []*pb.Booking
	for _, booking := range bookings {
		pbBookings = append(pbBookings, convertModelToPbBooking(booking))
	}

	return &pb.GetUserBookingsResponse{
		Bookings: pbBookings,
		Meta: &pb.PaginationMeta{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
	}, nil
}

// BookTickets implements the BookingService.BookTickets RPC
func (s *Server) BookTickets(ctx context.Context, req *pb.BookTicketsRequest) (*pb.Booking, error) {
	// Convert request to model
	bookingReq := &model.BookingRequest{
		ConcertID:   req.ConcertId,
		UserID:      req.UserId,
		TicketCount: int(req.TicketCount),
	}

	// Book tickets
	booking, err := s.bookingService.BookTickets(ctx, bookingReq)
	if err != nil {
		s.logger.Error("Failed to book tickets: %v", err)
		return nil, err
	}

	return convertModelToPbBooking(booking), nil
}

// CancelBooking implements the BookingService.CancelBooking RPC
func (s *Server) CancelBooking(ctx context.Context, req *pb.CancelBookingRequest) (*pb.CancelBookingResponse, error) {
	err := s.bookingService.CancelBooking(ctx, req.Id, req.UserId)
	if err != nil {
		s.logger.Error("Failed to cancel booking: %v", err)
		return nil, err
	}

	return &pb.CancelBookingResponse{
		Message: "Booking cancelled successfully",
	}, nil
}

// Helper functions to convert between model and protobuf types

// convertModelToPbConcert converts a model.Concert to a pb.Concert
func convertModelToPbConcert(concert *model.Concert) *pb.Concert {
	return &pb.Concert{
		Id:               concert.ID,
		Name:             concert.Name,
		Artist:           concert.Artist,
		Venue:            concert.Venue,
		ConcertDate:      timestamppb.New(concert.ConcertDate),
		TotalTickets:     int32(concert.TotalTickets),
		AvailableTickets: int32(concert.AvailableTickets),
		Price:            concert.Price,
		BookingStartTime: timestamppb.New(concert.BookingStartTime),
		BookingEndTime:   timestamppb.New(concert.BookingEndTime),
		Version:          int32(concert.Version),
		CreatedAt:        timestamppb.New(concert.CreatedAt),
		UpdatedAt:        timestamppb.New(concert.UpdatedAt),
	}
}

// convertModelToPbBooking converts a model.Booking to a pb.Booking
func convertModelToPbBooking(booking *model.Booking) *pb.Booking {
	return &pb.Booking{
		Id:          booking.ID,
		ConcertId:   booking.ConcertID,
		UserId:      booking.UserID,
		TicketCount: int32(booking.TicketCount),
		BookingTime: timestamppb.New(booking.BookingTime),
		Status:      string(booking.Status),
		CreatedAt:   timestamppb.New(booking.CreatedAt),
		UpdatedAt:   timestamppb.New(booking.UpdatedAt),
	}
}
