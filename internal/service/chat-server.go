package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	store "github.com/amirhlashgari/snapp-chat/pkg/nats"
	pb "github.com/amirhlashgari/snapp-chat/proto"

	"github.com/google/uuid"
)

type ChatService struct {
	pb.UnimplementedChatServiceServer
	store *store.JetStreamStore
	rooms map[string]*pb.ChatRoom
	users map[string]*pb.User
	mu    sync.RWMutex
}

func NewChatService(store *store.JetStreamStore) *ChatService {
	return &ChatService{
		store: store,
		rooms: make(map[string]*pb.ChatRoom),
		users: make(map[string]*pb.User),
	}
}

func (s *ChatService) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users, err := s.store.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
	}

	if req.Filter != "" {
		filteredUsers := []*pb.User{}
		for _, user := range users {
			if strings.Contains(strings.ToLower(user.Username), strings.ToLower(req.Filter)) {
				filteredUsers = append(filteredUsers, user)
			}
		}
		return &pb.ListUsersResponse{Users: filteredUsers}, nil
	}

	return &pb.ListUsersResponse{Users: users}, nil
}

func (s *ChatService) ListRooms(ctx context.Context, req *pb.ListRoomsRequest) (*pb.ListRoomsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rooms, err := s.store.GetRooms()
	if err != nil || len(rooms) == 0 {
		// Create default rooms if none exist
		defaultRooms := []*pb.ChatRoom{
			{
				Id:          uuid.New().String(),
				Name:        "Snapp",
				Description: "Snapp discussion room",
				Members:     []string{},
				CreatedAt:   time.Now().Unix(),
			},
			{
				Id:          uuid.New().String(),
				Name:        "Quera",
				Description: "Quera chat room",
				Members:     []string{},
				CreatedAt:   time.Now().Unix(),
			},
		}

		for _, room := range defaultRooms {
			if err := s.store.SaveRoom(room); err != nil {
				log.Printf("Error saving default room: %v", err)
			}
			s.rooms[room.Id] = room
		}

		rooms = defaultRooms
	}

	// Apply filter if provided
	if req.Filter != "" {
		filteredRooms := []*pb.ChatRoom{}
		for _, room := range rooms {
			if strings.Contains(strings.ToLower(room.Name), strings.ToLower(req.Filter)) {
				filteredRooms = append(filteredRooms, room)
			}
		}
		return &pb.ListRoomsResponse{Rooms: filteredRooms}, nil
	}

	return &pb.ListRoomsResponse{Rooms: rooms}, nil
}

func (s *ChatService) JoinRoom(ctx context.Context, req *pb.JoinRoomRequest) (*pb.JoinRoomResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rooms, err := s.store.GetRooms()
	if err != nil {
		return &pb.JoinRoomResponse{
			Success: false,
			Error:   "Failed to retrieve rooms",
		}, nil
	}

	var room *pb.ChatRoom
	for _, r := range rooms {
		if r.Id == req.RoomId {
			room = r
			break
		}
	}

	if room == nil {
		return &pb.JoinRoomResponse{
			Success: false,
			Error:   "Room not found",
		}, nil
	}

	// Check if user is already in the room
	for _, member := range room.Members {
		if member == req.UserId {
			return &pb.JoinRoomResponse{
				Success: true,
				Room:    room,
			}, nil
		}
	}

	room.Members = append(room.Members, req.UserId)

	// Save updated room
	if err := s.store.SaveRoom(room); err != nil {
		return &pb.JoinRoomResponse{
			Success: false,
			Error:   "Failed to save room",
		}, nil
	}

	return &pb.JoinRoomResponse{
		Success: true,
		Room:    room,
	}, nil
}

func (s *ChatService) LeaveRoom(ctx context.Context, req *pb.LeaveRoomRequest) (*pb.LeaveRoomResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rooms, err := s.store.GetRooms()
	if err != nil {
		return &pb.LeaveRoomResponse{
			Success: false,
			Error:   "Failed to retrieve rooms",
		}, nil
	}

	var room *pb.ChatRoom
	for _, r := range rooms {
		if r.Id == req.RoomId {
			room = r
			break
		}
	}

	if room == nil {
		return &pb.LeaveRoomResponse{
			Success: false,
			Error:   "Room not found",
		}, nil
	}

	// Remove user from room
	newMembers := []string{}
	for _, member := range room.Members {
		if member != req.UserId {
			newMembers = append(newMembers, member)
		}
	}
	room.Members = newMembers

	if err := s.store.SaveRoom(room); err != nil {
		return &pb.LeaveRoomResponse{
			Success: false,
			Error:   "Failed to save room",
		}, nil
	}

	return &pb.LeaveRoomResponse{
		Success: true,
	}, nil
}
