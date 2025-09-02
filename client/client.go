package client

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	pb "github.com/Gealber/yellowstone-tritonone/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}

type Client struct {
	address    string
	token      string
	processSub func(*pb.SubscribeUpdate)
	accounts   []string
	owners     []string
	txnsPID    []string

	closed bool
}

func New(
	accounts []string,
	owners []string,
	txnsPID []string,
	processSub func(*pb.SubscribeUpdate),
) (*Client, error) {
	grpcAddr := os.Getenv("GRPC_ENDPOINT")
	if grpcAddr == "" {
		return nil, errors.New("empty GRPC_ENDPOINT environment variable")
	}

	token := os.Getenv("GRPC_TOKEN")
	if grpcAddr == "" {
		return nil, errors.New("empty GRPC_TOKEN environment variable")
	}

	return &Client{
		address:    grpcAddr,
		token:      token,
		processSub: processSub,
		accounts:   accounts,
		owners:     owners,
		txnsPID:    txnsPID,
	}, nil
}

func (c *Client) Run() error {
	conn, err := grpc_connect(c.address, true)
	if err != nil {
		return err
	}
	defer conn.Close()

	return c.grpc_subscribe(conn)
}

func (c *Client) Close() {
	c.closed = true
}

func grpc_connect(address string, plaintext bool) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	if plaintext {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		pool, _ := x509.SystemCertPool()
		creds := credentials.NewClientTLSFromCert(pool, "")
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	opts = append(opts, grpc.WithKeepaliveParams(kacp))

	log.Println("Starting grpc client, connecting to", address)
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Client) grpc_subscribe(conn *grpc.ClientConn) error {
	var err error
	client := pb.NewGeyserClient(conn)

	var subscription pb.SubscribeRequest

	if (len(c.accounts) + len(c.owners)) > 0 {
		subscription.Accounts = make(map[string]*pb.SubscribeRequestFilterAccounts)
		subscription.Accounts["account_sub"] = &pb.SubscribeRequestFilterAccounts{}

		if len(c.accounts) > 0 {
			subscription.Accounts["account_sub"].Account = c.accounts
		}

		if len(c.owners) > 0 {
			subscription.Accounts["account_sub"].Owner = c.owners
		}
	}

	if len(c.txnsPID) > 0 {
		subscription.Transactions = make(map[string]*pb.SubscribeRequestFilterTransactions)
		subscription.Transactions["transactions_sub"] = &pb.SubscribeRequestFilterTransactions{
			AccountInclude: c.txnsPID,
		}
	}

	subscriptionJson, err := json.Marshal(&subscription)
	if err != nil {
		return err
	}
	log.Printf("Subscription request: %s", string(subscriptionJson))

	// Set up the subscription request
	ctx := context.Background()
	if c.token != "" {
		md := metadata.New(map[string]string{"x-token": c.token})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	stream, err := client.Subscribe(ctx)
	if err != nil {
		return err
	}
	err = stream.Send(&subscription)
	if err != nil {
		return err
	}

	for !c.closed {
		resp, err := stream.Recv()
		if err != nil {
			return err
		}

		c.processSub(resp)
	}

	return nil
}
