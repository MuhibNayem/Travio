package service

import (
	"context"

	"github.com/MuhibNayem/Travio/server/services/operator/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/operator/internal/repository"
)

type VendorService struct {
	repo repository.VendorRepository
}

func NewVendorService(repo repository.VendorRepository) *VendorService {
	return &VendorService{repo: repo}
}

func (s *VendorService) Create(ctx context.Context, name, email, phone, address string, rate float64) (*domain.Vendor, error) {
	vendor := &domain.Vendor{
		Name:           name,
		ContactEmail:   email,
		ContactPhone:   phone,
		Address:        address,
		Status:         "active",
		CommissionRate: rate,
	}
	if err := s.repo.Create(ctx, vendor); err != nil {
		return nil, err
	}
	return vendor, nil
}

func (s *VendorService) Get(ctx context.Context, id string) (*domain.Vendor, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *VendorService) Update(ctx context.Context, id, name, email, phone, address, status string, rate float64) (*domain.Vendor, error) {
	vendor, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		vendor.Name = name
	}
	if email != "" {
		vendor.ContactEmail = email
	}
	if phone != "" {
		vendor.ContactPhone = phone
	}
	if address != "" {
		vendor.Address = address
	}
	if status != "" {
		vendor.Status = status
	}
	if rate >= 0 {
		vendor.CommissionRate = rate
	}

	if err := s.repo.Update(ctx, vendor); err != nil {
		return nil, err
	}
	return vendor, nil
}

func (s *VendorService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *VendorService) List(ctx context.Context, page, limit int) ([]*domain.Vendor, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	return s.repo.List(ctx, page, limit)
}
