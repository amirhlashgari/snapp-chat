package store

import (
	"fmt"
	"time"

	pb "github.com/amirhlashgari/snapp-chat/proto"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type JetStreamStore struct {
	js nats.JetStreamContext
}

func NewJetStreamStore(nc *nats.Conn) (*JetStreamStore, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream context: %v", err)
	}

	streams := map[string][]string{
		"MESSAGES": {"chat.messages.>"},
		"USERS":    {"chat.users.>"},
		"ROOMS":    {"chat.rooms.>"},
	}

	for stream, subjects := range streams {
		_, err := js.AddStream(&nats.StreamConfig{
			Name:     stream,
			Subjects: subjects,
			MaxAge:   24 * 7 * time.Hour,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create stream %s: %v", stream, err)
		}
	}

	return &JetStreamStore{js: js}, nil
}

func (s *JetStreamStore) SaveMessage(msg *pb.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("chat.messages.%s", msg.RoomId)
	_, err = s.js.Publish(subject, data)
	return err
}

func (s *JetStreamStore) GetMessages(roomID string, limit int) ([]*pb.Message, error) {
	var messages []*pb.Message

	sub, err := s.js.SubscribeSync(fmt.Sprintf("chat.messages.%s", roomID))
	if err != nil {
		return nil, err
	}
	defer sub.Unsubscribe()

	for i := 0; i < limit; i++ {
		msg, err := sub.NextMsg(time.Second)
		if err == nats.ErrTimeout {
			break
		}
		if err != nil {
			return nil, err
		}

		var pbMsg pb.Message
		if err := proto.Unmarshal(msg.Data, &pbMsg); err != nil {
			return nil, err
		}
		messages = append(messages, &pbMsg)
	}

	return messages, nil
}

func (s *JetStreamStore) SaveUser(user *pb.User) error {
	data, err := proto.Marshal(user)
	if err != nil {
		return err
	}

	_, err = s.js.Publish(fmt.Sprintf("chat.users.%s", user.Id), data)
	return err
}

func (s *JetStreamStore) GetUsers() ([]*pb.User, error) {
	var users []*pb.User

	sub, err := s.js.SubscribeSync("chat.users.>")
	if err != nil {
		return nil, err
	}
	defer sub.Unsubscribe()

	for {
		msg, err := sub.NextMsg(time.Second)
		if err == nats.ErrTimeout {
			break
		}
		if err != nil {
			return nil, err
		}

		var user pb.User
		if err := proto.Unmarshal(msg.Data, &user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

func (s *JetStreamStore) SaveRoom(room *pb.ChatRoom) error {
	data, err := proto.Marshal(room)
	if err != nil {
		return err
	}

	_, err = s.js.Publish(fmt.Sprintf("chat.rooms.%s", room.Id), data)
	return err
}

func (s *JetStreamStore) GetRooms() ([]*pb.ChatRoom, error) {
	var rooms []*pb.ChatRoom

	sub, err := s.js.SubscribeSync("chat.rooms.>")
	if err != nil {
		return nil, err
	}
	defer sub.Unsubscribe()

	for {
		msg, err := sub.NextMsg(time.Second)
		if err == nats.ErrTimeout {
			break
		}
		if err != nil {
			return nil, err
		}

		var room pb.ChatRoom
		if err := proto.Unmarshal(msg.Data, &room); err != nil {
			return nil, err
		}
		rooms = append(rooms, &room)
	}

	return rooms, nil
}
