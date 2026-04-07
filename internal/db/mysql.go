package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	MaxOpen  int
	MaxIdle  int
}

func (c *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		c.Username, c.Password, c.Host, c.Port, c.Database)
}

func (c *MySQLConfig) ServerDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True",
		c.Username, c.Password, c.Host, c.Port)
}

type MySQL struct {
	db *sql.DB
}

func NewMySQL(cfg *MySQLConfig) (*MySQL, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, err
	}

	if cfg.MaxOpen > 0 {
		db.SetMaxOpenConns(cfg.MaxOpen)
	}
	if cfg.MaxIdle > 0 {
		db.SetMaxIdleConns(cfg.MaxIdle)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &MySQL{db: db}, nil
}

func EnsureDatabase(cfg *MySQLConfig) error {
	serverDB, err := sql.Open("mysql", cfg.ServerDSN())
	if err != nil {
		return err
	}
	defer serverDB.Close()

	if err := serverDB.Ping(); err != nil {
		return err
	}

	_, err = serverDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.Database))
	return err
}

func (m *MySQL) Close() error {
	return m.db.Close()
}

func (m *MySQL) DB() *sql.DB {
	return m.db
}
