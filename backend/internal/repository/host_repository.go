package repository

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/kineticops/backend/internal/models"
)

type HostRepository struct {
	*BaseRepository
}

func NewHostRepository(base *BaseRepository) *HostRepository {
	return &HostRepository{BaseRepository: base}
}

func (r *HostRepository) AddHost(ctx context.Context, host *models.Host) error {
	query := `INSERT INTO hosts (id, user_id, name, ip_address, description, status, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7)`
	err := r.Exec(ctx, query, host.ID, host.UserID, host.Name, host.IPAddress, host.Description, host.Status, host.CreatedAt)
	if err != nil {
		log.Printf("❌ AddHost error: %v", err)
		return err
	}
	return nil
}

func (r *HostRepository) GetHostByID(ctx context.Context, id string) (*models.Host, error) {
	var host models.Host
	query := `SELECT id, user_id, name, ip_address, description, status, created_at FROM hosts WHERE id = $1`
	err := r.DB.QueryRow(ctx, query, id).Scan(&host.ID, &host.UserID, &host.Name, &host.IPAddress, &host.Description, &host.Status, &host.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		log.Printf("❌ GetHostByID error: %v", err)
		return nil, err
	}
	return &host, nil
}

func (r *HostRepository) GetHostsByUser(ctx context.Context, userID string) ([]*models.Host, error) {
	rows, err := r.DB.Query(ctx, `SELECT id, user_id, name, ip_address, description, status, created_at 
	                              FROM hosts WHERE user_id = $1`, userID)
	if err != nil {
		log.Printf("❌ GetHostsByUser error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var hosts []*models.Host
	for rows.Next() {
		var h models.Host
		err := rows.Scan(&h.ID, &h.UserID, &h.Name, &h.IPAddress, &h.Description, &h.Status, &h.CreatedAt)
		if err != nil {
			log.Printf("❌ Scan host error: %v", err)
			continue
		}
		hosts = append(hosts, &h)
	}
	return hosts, nil
}

func (r *HostRepository) UpdateHost(ctx context.Context, host *models.Host) error {
	query := `UPDATE hosts SET user_id = $1, name = $2, ip_address = $3, description = $4, status = $5 WHERE id = $6`
	err := r.Exec(ctx, query, host.UserID, host.Name, host.IPAddress, host.Description, host.Status, host.ID)
	if err != nil {
		log.Printf("❌ UpdateHost error: %v", err)
		return err
	}
	return nil
}

func (r *HostRepository) DeleteHost(ctx context.Context, id string) error {
	query := `DELETE FROM hosts WHERE id = $1`
	err := r.Exec(ctx, query, id)
	if err != nil {
		log.Printf("❌ DeleteHost error: %v", err)
		return err
	}
	return nil
}
