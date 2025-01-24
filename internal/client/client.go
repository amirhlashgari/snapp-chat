package client

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/amirhlashgari/snapp-chat/proto"
	"github.com/google/uuid"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	userID      string
	username    string
	nc          *nats.Conn
	js          nats.JetStreamContext
	currentRoom *pb.ChatRoom
	service     pb.ChatServiceClient
	msgChan     chan *pb.Message
	done        chan struct{}
	mu          sync.RWMutex
}

func NewClient(userID, username string, nc *nats.Conn, service pb.ChatServiceClient) (*Client, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream context: %v", err)
	}

	client := &Client{
		userID:   userID,
		username: username,
		nc:       nc,
		js:       js,
		service:  service,
		msgChan:  make(chan *pb.Message, 100),
		done:     make(chan struct{}),
	}

	// Save/update user in store
	user := &pb.User{
		Id:       userID,
		Username: username,
		Status:   "online",
		LastSeen: time.Now().Unix(),
	}

	data, err := proto.Marshal(user)
	if err != nil {
		return nil, err
	}

	if _, err := js.Publish("chat.users."+userID, data); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) ListUsers(filter string) ([]*pb.User, error) {
	resp, err := c.service.ListUsers(context.Background(), &pb.ListUsersRequest{
		Filter: filter,
	})
	if err != nil {
		return nil, err
	}
	return resp.Users, nil
}

func (c *Client) ListRooms(filter string) ([]*pb.ChatRoom, error) {
	resp, err := c.service.ListRooms(context.Background(), &pb.ListRoomsRequest{
		Filter: filter,
	})
	if err != nil {
		return nil, err
	}
	return resp.Rooms, nil
}

func (c *Client) JoinRoom(roomID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	resp, err := c.service.JoinRoom(context.Background(), &pb.JoinRoomRequest{
		RoomId: roomID,
		UserId: c.userID,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("failed to join room: %s", resp.Error)
	}

	// Unsubscribe from previous room if any
	if c.currentRoom != nil {
		c.LeaveRoom(c.currentRoom.Id)
	}

	c.currentRoom = resp.Room
	sub, err := c.js.Subscribe(
		fmt.Sprintf("chat.messages.%s", roomID),
		func(msg *nats.Msg) {
			var pbMsg pb.Message
			if err := proto.Unmarshal(msg.Data, &pbMsg); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				return
			}
			c.msgChan <- &pbMsg
		},
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to room: %v", err)
	}

	go func() {
		<-c.done
		sub.Unsubscribe()
	}()

	return nil
}

func (c *Client) LeaveRoom(roomID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	resp, err := c.service.LeaveRoom(context.Background(), &pb.LeaveRoomRequest{
		RoomId: roomID,
		UserId: c.userID,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("failed to leave room: %s", resp.Error)
	}

	close(c.done)
	c.done = make(chan struct{})
	c.currentRoom = nil
	return nil
}

func (c *Client) SendMessage(content string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.currentRoom == nil {
		return fmt.Errorf("not in any room")
	}

	msg := &pb.Message{
		Id:        uuid.New().String(),
		RoomId:    c.currentRoom.Id,
		UserId:    c.userID,
		Username:  c.username,
		Content:   content,
		Timestamp: time.Now().Unix(),
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = c.js.Publish(fmt.Sprintf("chat.messages.%s", c.currentRoom.Id), data)
	return err
}

func (c *Client) MessageChannel() <-chan *pb.Message {
	return c.msgChan
}

func (c *Client) Close() error {
	if c.currentRoom != nil {
		if err := c.LeaveRoom(c.currentRoom.Id); err != nil {
			return err
		}
	}

	user := &pb.User{
		Id:       c.userID,
		Username: c.username,
		Status:   "offline",
		LastSeen: time.Now().Unix(),
	}

	data, err := proto.Marshal(user)
	if err != nil {
		return err
	}

	if _, err := c.js.Publish("chat.users."+c.userID, data); err != nil {
		return err
	}

	c.nc.Close()
	return nil
}
