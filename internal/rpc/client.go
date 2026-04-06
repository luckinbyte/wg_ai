package rpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ss "github.com/luckinbyte/wg_ai/proto/ss"
)

type ClientConfig struct {
	DBAddr    string
	LoginAddr string
}

type Client struct {
	dbConn    *grpc.ClientConn
	dbClient  ss.DBServiceClient
	loginConn *grpc.ClientConn
	login     ss.LoginServiceClient
}

func NewClient(cfg *ClientConfig) *Client {
	return &Client{}
}

func (c *Client) ConnectDB(addr string) error {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	c.dbConn = conn
	c.dbClient = ss.NewDBServiceClient(conn)
	return nil
}

func (c *Client) ConnectLogin(addr string) error {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	c.loginConn = conn
	c.login = ss.NewLoginServiceClient(conn)
	return nil
}

func (c *Client) Close() {
	if c.dbConn != nil {
		c.dbConn.Close()
	}
	if c.loginConn != nil {
		c.loginConn.Close()
	}
}

func (c *Client) LoadRole(ctx context.Context, rid int64) ([]byte, bool, error) {
	resp, err := c.dbClient.LoadRole(ctx, &ss.LoadRoleRequest{Rid: rid})
	if err != nil {
		return nil, false, err
	}
	return resp.Data, resp.Found, nil
}

func (c *Client) SaveRole(ctx context.Context, rid int64, data []byte) error {
	_, err := c.dbClient.SaveRole(ctx, &ss.SaveRoleRequest{Rid: rid, Data: data})
	return err
}

func (c *Client) CreateUser(ctx context.Context, username, passwordHash string) (int64, error) {
	resp, err := c.dbClient.CreateUser(ctx, &ss.CreateUserRequest{
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return 0, err
	}
	return resp.Uid, nil
}

func (c *Client) GetUser(ctx context.Context, username string) (int64, string, bool, error) {
	resp, err := c.dbClient.GetUser(ctx, &ss.GetUserRequest{Username: username})
	if err != nil {
		return 0, "", false, err
	}
	return resp.Uid, resp.PasswordHash, resp.Found, nil
}

func (c *Client) ValidateToken(ctx context.Context, token string) (int64, bool, error) {
	resp, err := c.login.ValidateToken(ctx, &ss.ValidateTokenRequest{Token: token})
	if err != nil {
		return 0, false, err
	}
	return resp.Uid, resp.Valid, nil
}

func (c *Client) NotifyLogin(ctx context.Context, uid int64, token string) (bool, error) {
	resp, err := c.login.NotifyLogin(ctx, &ss.LoginNotifyRequest{Uid: uid, Token: token})
	if err != nil {
		return false, err
	}
	return resp.Success, nil
}
