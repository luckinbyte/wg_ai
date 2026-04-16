package db

import (
	"context"
	"encoding/json"
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
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey)
		if err == nil {
			return &ss.LoadRoleResponse{Data: cached, Found: true}, nil
		}
	}

	// Query MySQL
	var data []byte
	err := s.mysql.db.QueryRowContext(ctx,
		"SELECT data FROM role WHERE rid = ?", req.Rid).Scan(&data)
	if err != nil {
		return &ss.LoadRoleResponse{Found: false}, nil
	}

	// Cache it
	if s.redis != nil {
		_ = s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	}

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
	if s.redis != nil {
		_ = s.redis.Set(ctx, cacheKey, req.Data, 5*time.Minute)
	}

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

func (s *DBService) LoadAllCities(req *ss.LoadAllCitiesRequest, stream ss.DBService_LoadAllCitiesServer) error {
	rows, err := s.mysql.db.QueryContext(stream.Context(),
		"SELECT rid, data FROM role")
	if err != nil {
		return fmt.Errorf("query roles: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rid int64
		var data []byte
		if err := rows.Scan(&rid, &data); err != nil {
			continue
		}

		// 解析序列化的玩家数据，提取 city
		var raw struct {
			Arrays struct {
				City json.RawMessage `json:"city"`
			} `json:"arrays"`
		}
		if err := json.Unmarshal(data, &raw); err != nil || raw.Arrays.City == nil {
			continue
		}

		var city struct {
			Position struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			} `json:"position"`
			CityID    int64           `json:"city_id"`
			Buildings json.RawMessage `json:"buildings"`
		}
		if err := json.Unmarshal(raw.Arrays.City, &city); err != nil {
			continue
		}

		if err := stream.Send(&ss.CityDataMsg{
			Rid:       rid,
			CityId:    city.CityID,
			PosX:      city.Position.X,
			PosY:      city.Position.Y,
			Buildings: city.Buildings,
		}); err != nil {
			return err
		}
	}
	return nil
}
