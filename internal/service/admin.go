package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"inventariskantor/internal/models"
)

// ---------- Users ----------

func (s *Service) ListUsers(ctx context.Context) ([]models.UserRow, error) {
	rows, err := s.Pool.Query(ctx, `SELECT id, name, email, role, created_at FROM users ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.UserRow
	for rows.Next() {
		var u models.UserRow
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *Service) CreateUser(ctx context.Context, name, email, password, role string) error {
	if len(password) < 8 {
		return fmt.Errorf("password minimal 8 karakter")
	}
	if role != "admin" && role != "petugas" {
		role = "petugas"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.Pool.Exec(ctx,
		`INSERT INTO users (name, email, password_hash, role) VALUES ($1,$2,$3,$4)
		 ON CONFLICT (email) DO UPDATE SET name=$1, password_hash=$3, role=$4`,
		name, email, string(hash), role)
	return err
}

func (s *Service) UpdateUserRole(ctx context.Context, email, role string) error {
	if role != "admin" && role != "petugas" {
		return fmt.Errorf("role tidak valid")
	}
	_, err := s.Pool.Exec(ctx, `UPDATE users SET role=$2 WHERE email=$1`, email, role)
	return err
}

func (s *Service) DeleteUser(ctx context.Context, email string) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM users WHERE email=$1`, email)
	return err
}

// ChangePassword: user ganti password sendiri (verifikasi password lama).
func (s *Service) ChangePassword(ctx context.Context, email, oldPW, newPW string) error {
	if len(newPW) < 8 {
		return fmt.Errorf("password baru minimal 8 karakter")
	}
	var hash string
	err := s.Pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE email=$1`, email).Scan(&hash)
	if err != nil {
		return fmt.Errorf("user tidak ditemukan")
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(oldPW)) != nil {
		return fmt.Errorf("password lama salah")
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPW), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.Pool.Exec(ctx, `UPDATE users SET password_hash=$2 WHERE email=$1`, email, string(newHash))
	return err
}

// ---------- Categories & Locations CRUD ----------

func (s *Service) CreateCategory(ctx context.Context, name string) error {
	id := Slugify(name)
	if id == "" {
		return fmt.Errorf("nama tidak valid")
	}
	_, err := s.Pool.Exec(ctx,
		`INSERT INTO categories (id, name) VALUES ($1,$2) ON CONFLICT (id) DO UPDATE SET name=$2`, id, name)
	return err
}

func (s *Service) RenameCategory(ctx context.Context, id, newName string) error {
	var oldName string
	if err := s.Pool.QueryRow(ctx, `SELECT name FROM categories WHERE id=$1`, id).Scan(&oldName); err != nil {
		return fmt.Errorf("kategori tidak ditemukan")
	}
	if _, err := s.Pool.Exec(ctx, `UPDATE categories SET name=$2 WHERE id=$1`, id, newName); err != nil {
		return err
	}
	if oldName != newName {
		_, err := s.Pool.Exec(ctx, `UPDATE inventory_items SET category=$2 WHERE category=$1`, oldName, newName)
		return err
	}
	return nil
}

func (s *Service) DeleteCategory(ctx context.Context, id string) error {
	var name string
	if err := s.Pool.QueryRow(ctx, `SELECT name FROM categories WHERE id=$1`, id).Scan(&name); err != nil {
		return fmt.Errorf("kategori tidak ditemukan")
	}
	var used int
	_ = s.Pool.QueryRow(ctx, `SELECT count(*) FROM inventory_items WHERE category=$1`, name).Scan(&used)
	if used > 0 {
		return fmt.Errorf("kategori masih dipakai %d barang", used)
	}
	_, err := s.Pool.Exec(ctx, `DELETE FROM categories WHERE id=$1`, id)
	return err
}

func (s *Service) CreateLocation(ctx context.Context, name string) error {
	id := Slugify(name)
	if id == "" {
		return fmt.Errorf("nama tidak valid")
	}
	_, err := s.Pool.Exec(ctx,
		`INSERT INTO locations (id, name) VALUES ($1,$2) ON CONFLICT (id) DO UPDATE SET name=$2`, id, name)
	return err
}

func (s *Service) RenameLocation(ctx context.Context, id, newName string) error {
	var oldName string
	if err := s.Pool.QueryRow(ctx, `SELECT name FROM locations WHERE id=$1`, id).Scan(&oldName); err != nil {
		return fmt.Errorf("lokasi tidak ditemukan")
	}
	if _, err := s.Pool.Exec(ctx, `UPDATE locations SET name=$2 WHERE id=$1`, id, newName); err != nil {
		return err
	}
	if oldName != newName {
		_, _ = s.Pool.Exec(ctx, `UPDATE inventory_items SET location=$2 WHERE location=$1`, oldName, newName)
		_, _ = s.Pool.Exec(ctx, `UPDATE inventory_items i SET location_id = l.id FROM locations l WHERE i.location=l.name`)
	}
	return nil
}

func (s *Service) DeleteLocation(ctx context.Context, id string) error {
	var name string
	if err := s.Pool.QueryRow(ctx, `SELECT name FROM locations WHERE id=$1`, id).Scan(&name); err != nil {
		return fmt.Errorf("lokasi tidak ditemukan")
	}
	var used int
	_ = s.Pool.QueryRow(ctx, `SELECT count(*) FROM inventory_items WHERE location=$1`, name).Scan(&used)
	if used > 0 {
		return fmt.Errorf("lokasi masih dipakai %d barang", used)
	}
	_, err := s.Pool.Exec(ctx, `DELETE FROM locations WHERE id=$1`, id)
	return err
}

// ---------- Scheduler ----------

// MonthlyPDFReport generates the PDF report; returns bytes (called by cron).
func (s *Service) MonthlyReportLog(ctx context.Context) error {
	slog.Info("monthly report triggered")
	return nil
}

// RunDriveImport runs a Drive CSV import cycle if configured (cron).
func (s *Service) RunDriveImport(ctx context.Context) (string, error) {
	slog.Info("scheduled drive import triggered")
	return "not configured", nil
}

var _ = time.Now
