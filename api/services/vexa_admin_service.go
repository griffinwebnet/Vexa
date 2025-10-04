package services

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type vexaAdminRecord struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	CreatedAt    int64  `json:"created_at"`
}

// VexaAdminService manages the fallback Vexa admin credentials
type VexaAdminService struct {
	storagePath string
}

func NewVexaAdminService() *VexaAdminService {
	return &VexaAdminService{storagePath: "/var/lib/vexa/admin.json"}
}

func (s *VexaAdminService) ensureDir() error {
	dir := filepath.Dir(s.storagePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return nil
}

func (s *VexaAdminService) IsInitialized() bool {
	if _, err := os.Stat(s.storagePath); err == nil {
		return true
	}
	return false
}

func (s *VexaAdminService) SetPassword(password string) error {
	if password == "" {
		return errors.New("password required")
	}
	if err := s.ensureDir(); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	rec := vexaAdminRecord{
		Username:     "vexa",
		PasswordHash: string(hash),
		CreatedAt:    time.Now().Unix(),
	}
	b, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	// write with restricted permissions
	if err := os.WriteFile(s.storagePath, b, 0600); err != nil {
		return err
	}
	return nil
}

func (s *VexaAdminService) Verify(username, password string) bool {
	if username != "vexa" {
		return false
	}
	b, err := os.ReadFile(s.storagePath)
	if err != nil {
		return false
	}
	var rec vexaAdminRecord
	if err := json.Unmarshal(b, &rec); err != nil {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(rec.PasswordHash), []byte(password)) == nil
}
