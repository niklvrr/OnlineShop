package client

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	paymentpb "github.com/niklvrr/OnlineShop/proto-contracts/gen/go/payment"
)

type PaymentClient struct {
	conn   *grpc.ClientConn
	client paymentpb.PaymentServiceClient
}

func NewPaymentClient(url string) (*PaymentClient, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &PaymentClient{
		conn:   conn,
		client: paymentpb.NewPaymentServiceClient(conn),
	}, nil
}

func (c *PaymentClient) InitiatePayment(ctx context.Context, req *paymentpb.InitiatePaymentRequest) (*paymentpb.InitiatePaymentResponse, error) {
	return c.client.InitiatePayment(ctx, req)
}

func (c *PaymentClient) GetPaymentStatus(ctx context.Context, req *paymentpb.GetPaymentStatusRequest) (*paymentpb.GetPaymentStatusResponse, error) {
	return c.client.GetPaymentStatus(ctx, req)
}

func (c *PaymentClient) Close() error {
	return c.conn.Close()
}

