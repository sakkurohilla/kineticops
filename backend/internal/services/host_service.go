package services

import (
	"context"
	"kineticops/backend/internal/repo"
	"kineticops/backend/models"
)

type HostService struct{ Repo *repo.HostRepo }

func (s *HostService) Create(ctx context.Context, h *models.Host) error { return s.Repo.Create(ctx, h) }
func (s *HostService) List(ctx context.Context, owner int) ([]models.Host, error) {
	return s.Repo.List(ctx, owner)
}
func (s *HostService) Update(ctx context.Context, id int, name, ip string, owner int) error {
	return s.Repo.Update(ctx, id, name, ip, owner)
}
func (s *HostService) Delete(ctx context.Context, id, owner int) error {
	return s.Repo.Delete(ctx, id, owner)
}
func (s *HostService) Heartbeat(ctx context.Context, id, owner int) error {
	return s.Repo.Heartbeat(ctx, id, owner)
}
