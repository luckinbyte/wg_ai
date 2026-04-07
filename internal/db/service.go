package db

import (
	"context"
	"fmt"
	"time"

	ss "github.com/luckinbyte/wg_ai/proto/ss"
)

type DBService struct {
	ss.UnimplementedDBServiceServer
	mysql *MySQL
	redis *Redis
}

func NewDBService(mysql *MySQL, redis *Redis) *DBService {
	return &DBService{
		mysql: mysql,
		redis: redis,
	}
}

func EnsureTables(mysql *MySQL) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS user (
			uid BIGINT PRIMARY KEY AUTO_INCREMENT,
			username VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS role (
			rid BIGINT PRIMARY KEY,
			data LONGBLOB NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
	}

	for _, stmt := range stmts {
		if _, err := mysql.db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}

func (s *DBService) LoadRole(ctx context.Context, req *ss.LoadRoleRequest) (*ss.LoadRoleResponse, error) {
	cacheKey := fmt.Sprintf("role:%d", req.Rid)

	// Try Redis first
	cached, err := s.redis.Get(ctx, cacheKey)
	if err == nil {
		return &ss.LoadRoleResponse{Data: cached, Found: true}, nil
	}

	// Query MySQL
	var data []byte
	err = s.mysql.db.QueryRowContext(ctx,
		"SELECT data FROM role WHERE rid = ?", req.Rid).Scan(&data)
	if err != nil {
		return &ss.LoadRoleResponse{Found: false}, nil
	}

	// Cache it
	_ = s.redis.Set(ctx, cacheKey, data, 5*time.Minute)

	return &ss.LoadRoleResponse{Data: data, Found: true}, nil
}

func (s *DBService) SaveRole(ctx context.Context, req *ss.SaveRoleRequest) (*ss.SaveRoleResponse, error) {
	_, err := s.mysql.db.ExecContext(ctx,
		"INSERT INTO role (rid, data, updated_at) VALUES (?, ?, NOW()) "+
			"ON DUPLICATE KEY UPDATE data = ?, updated_at = NOW()",
		req.Rid, req.Data, req.Data)
	if err != nil {
		return &ss.SaveRoleResponse{Success: false}, nil
	}

	cacheKey := fmt.Sprintf("role:%d", req.Rid)
	_ = s.redis.Set(ctx, cacheKey, req.Data, 5*time.Minute)

	return &ss.SaveRoleResponse{Success: true}, nil
}

func (s *DBService) CreateUser(ctx context.Context, req *ss.CreateUserRequest) (*ss.CreateUserResponse, error) {
	result, err := s.mysql.db.ExecContext(ctx,
		"INSERT INTO user (username, password_hash, created_at) VALUES (?, ?, NOW())",
		req.Username, req.PasswordHash)
	if err != nil {
		return &ss.CreateUserResponse{Uid: 0}, nil
	}

	uid, _ := result.LastInsertId()
	return &ss.CreateUserResponse{Uid: uid}, nil
}

func (s *DBService) GetUser(ctx context.Context, req *ss.GetUserRequest) (*ss.GetUserResponse, error) {
	var uid int64
	var passwordHash string
	err := s.mysql.db.QueryRowContext(ctx,
		"SELECT uid, password_hash FROM user WHERE username = ?", req.Username).
		Scan(&uid, &passwordHash)
	if err != nil {
		return &ss.GetUserResponse{Found: false}, nil
	}
	return &ss.GetUserResponse{Uid: uid, PasswordHash: passwordHash, Found: true}, nil
}
