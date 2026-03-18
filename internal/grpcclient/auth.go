package grpcclient

import (
	"context"
	"fmt"
	pb "gkeeper/api/proto"

	"google.golang.org/protobuf/proto"
)

func (c *Client) Login(ctx context.Context, email string, password string) error {
	request, reqErr := c.client.Login(ctx, pb.LoginRequest_builder{
		Email:    proto.String(email),
		Password: proto.String(password),
	}.Build())

	if reqErr != nil {
		return reqErr
	}

	token := request.GetResult()
	c.SetToken(token)

	fmt.Println("Successfully logged in")
	return nil
}

func (c *Client) Register(ctx context.Context, email string, password string) error {
	request, reqErr := c.client.Register(ctx, pb.RegisterRequest_builder{
		Email:    proto.String(email),
		Password: proto.String(password),
	}.Build())

	if reqErr != nil {
		return reqErr
	}

	fmt.Println(request.GetResult())
	return nil
}
