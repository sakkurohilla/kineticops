package repo

import (
	"context"

	"kineticops/backend/models"

	"github.com/jmoiron/sqlx"
)

type HostRepo struct{ DB *sqlx.DB }

func (r *HostRepo) Create(ctx context.Context, h *models.Host) error {
	q := `INSERT INTO hosts (name, ip_address, owner_id)
	       VALUES ($1,$2,$3)
	       RETURNING id,name,ip_address,status,last_seen,owner_id,created_at,updated_at`
	return r.DB.QueryRowContext(ctx, q, h.Name, h.IPAddress, h.OwnerID).
		Scan(&h.ID, &h.Name, &h.IPAddress, &h.Status, &h.LastSeen, &h.OwnerID, &h.CreatedAt, &h.UpdatedAt)
}
func (r *HostRepo) List(ctx context.Context, owner int) ([]models.Host, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,name,ip_address,status,last_seen,owner_id,created_at,updated_at 
		FROM hosts WHERE owner_id=$1 ORDER BY id DESC`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Host
	for rows.Next() {
		var h models.Host
		_ = rows.Scan(&h.ID, &h.Name, &h.IPAddress, &h.Status, &h.LastSeen, &h.OwnerID, &h.CreatedAt, &h.UpdatedAt)
		out = append(out, h)
	}
	return out, nil
}

func (r *HostRepo) Update(ctx context.Context, id int, name string, ip string, owner int) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE hosts SET name=$1, ip_address=$2, updated_at=NOW()
		WHERE id=$3 AND owner_id=$4`, name, ip, id, owner)
	return err
}

func (r *HostRepo) Delete(ctx context.Context, id int, owner int) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM hosts WHERE id=$1 AND owner_id=$2`, id, owner)
	return err
}

func (r *HostRepo) Heartbeat(ctx context.Context, id int, owner int) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE hosts SET last_seen=NOW(), status='online'
		WHERE id=$1 AND owner_id=$2`, id, owner)
	return err
}
