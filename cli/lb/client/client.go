package client

import (
	"context"
	"fmt"
	lbb "github.com/ppal31/grpc-lab/cli/lb/client/balancer"
	chatv1 "github.com/ppal31/grpc-lab/generated/chat/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
	"io"
	"log"
	"math/rand"
)

type Client struct {
	accountIds []string
	zkAddrs    []string // Address of server to connect to
}

func (c *Client) Ping() error {
	cc, closer, err := c.setupClient()
	if err != nil {
		return err
	}
	defer closer.Close()

	for i := 0; i < 1000; i++ {
		accountId := c.accountIds[rand.Intn(len(c.accountIds))]
		ctx := context.WithValue(context.Background(), "accountId", accountId)
		r, err := cc.Ping(ctx, &chatv1.PingRequest{Message: "PING"})
		if err != nil {
			return err
		}
		log.Printf("Reply : %s", r.GetMessage())
	}
	return nil
}

func (c *Client) Chat(accountId string) error {
	cc, closer, err := c.setupClient()
	if err != nil {
		return err
	}
	defer closer.Close()
	ctx := context.WithValue(context.Background(), "accountId", accountId)
	cs, err := cc.Chat(ctx)
	if err != nil {
		return err
	}
	done := make(chan struct{})

	go func() {
		for {
			cm, err := cs.Recv()
			if err == io.EOF {
				done <- struct{}{}
				return
			}
			if err != nil {
				log.Fatalf(err.Error())
			}
			log.Printf(cm.Message)
		}

	}()

	for i := 0; i < 100; i++ {
		err := cs.Send(&chatv1.ChatMessage{Message: fmt.Sprintf("Message %d", i)})
		if err != nil {
			return err
		}
	}
	cs.CloseSend()
	<-done
	return nil
}

func (c *Client) setupClient() (chatv1.ChatServiceClient, io.Closer, error) {
	balancer.Register(lbb.NewBuilder())
	rb, err := lbb.NewZkBuilder(c.zkAddrs)
	if err != nil {
		return nil, nil, err
	}
	resolver.Register(rb)
	conn, err := grpc.Dial(fmt.Sprintf("%s:///", rb.Scheme()), grpc.WithInsecure(), grpc.WithBalancerName(lbb.Name))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	cc := chatv1.NewChatServiceClient(conn)
	return cc, conn, nil
}

func NewClient(zkAddrs, accountIds []string) *Client {
	return &Client{zkAddrs: zkAddrs, accountIds: accountIds}
}
