package handler

import (
	"context"

	"github.com/MuhibNayem/Travio/server/services/operator/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/operator/internal/service"

	pb "github.com/MuhibNayem/Travio/server/api/proto/operator/v1"
)

type GrpcHandler struct {
	pb.UnimplementedVendorServiceServer
	service *service.VendorService
}

func NewGrpcHandler(service *service.VendorService) *GrpcHandler {
	return &GrpcHandler{service: service}
}

func (h *GrpcHandler) CreateVendor(ctx context.Context, req *pb.CreateVendorRequest) (*pb.CreateVendorResponse, error) {
	v, err := h.service.Create(ctx, req.Name, req.ContactEmail, req.ContactPhone, req.Address, req.CommissionRate)
	if err != nil {
		return nil, err
	}
	return &pb.CreateVendorResponse{Vendor: mapDomainToProto(v)}, nil
}

func (h *GrpcHandler) GetVendor(ctx context.Context, req *pb.GetVendorRequest) (*pb.GetVendorResponse, error) {
	v, err := h.service.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetVendorResponse{Vendor: mapDomainToProto(v)}, nil
}

func (h *GrpcHandler) UpdateVendor(ctx context.Context, req *pb.UpdateVendorRequest) (*pb.UpdateVendorResponse, error) {
	v, err := h.service.Update(ctx, req.Id, req.Name, req.ContactEmail, req.ContactPhone, req.Address, req.Status, req.CommissionRate)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateVendorResponse{Vendor: mapDomainToProto(v)}, nil
}

func (h *GrpcHandler) ListVendors(ctx context.Context, req *pb.ListVendorsRequest) (*pb.ListVendorsResponse, error) {
	vendors, total, err := h.service.List(ctx, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, err
	}

	var pbVendors []*pb.Vendor
	for _, v := range vendors {
		pbVendors = append(pbVendors, mapDomainToProto(v))
	}

	return &pb.ListVendorsResponse{
		Vendors: pbVendors,
		Total:   total,
		Page:    req.Page,
		Limit:   req.Limit,
	}, nil
}

func (h *GrpcHandler) DeleteVendor(ctx context.Context, req *pb.DeleteVendorRequest) (*pb.DeleteVendorResponse, error) {
	err := h.service.Delete(ctx, req.Id)
	if err != nil {
		return &pb.DeleteVendorResponse{Success: false}, err
	}
	return &pb.DeleteVendorResponse{Success: true}, nil
}

func mapDomainToProto(v *domain.Vendor) *pb.Vendor {
	return &pb.Vendor{
		Id:             v.ID,
		Name:           v.Name,
		ContactEmail:   v.ContactEmail,
		ContactPhone:   v.ContactPhone,
		Address:        v.Address,
		Status:         v.Status,
		CommissionRate: v.CommissionRate,
		CreatedAt:      v.CreatedAt.String(),
		UpdatedAt:      v.UpdatedAt.String(),
	}
}
