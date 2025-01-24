package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	store "github.com/amirhlashgari/snapp-chat/pkg/nats"
	pb "github.com/amirhlashgari/snapp-chat/proto"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

func setupTestService(t *testing.T) (*ChatService, *nats.Conn) {
	nc, err := nats.Connect(nats.DefaultURL)
	require.NoError(t, err)

	jetStreamStore, err := store.NewJetStreamStore(nc)
	require.NoError(t, err)

	return NewChatService(jetStreamStore), nc
}

func TestJoinRoom(t *testing.T) {
	service, nc := setupTestService(t)
	defer nc.Close()

	// Get an existing room
	roomsResp, err := service.ListRooms(context.Background(), &pb.ListRoomsRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, roomsResp.Rooms)

	testRoom := roomsResp.Rooms[0]
	testUserID := uuid.New().String()

	// Join room
	joinResp, err := service.JoinRoom(context.Background(), &pb.JoinRoomRequest{
		RoomId: testRoom.Id,
		UserId: testUserID,
	})
	assert.NoError(t, err)
	assert.True(t, joinResp.Success)
	assert.Contains(t, joinResp.Room.Members, testUserID)

	// Try joining again (should still succeed)
	secondJoinResp, err := service.JoinRoom(context.Background(), &pb.JoinRoomRequest{
		RoomId: testRoom.Id,
		UserId: testUserID,
	})
	assert.NoError(t, err)
	assert.True(t, secondJoinResp.Success)
}

func TestLeaveRoom(t *testing.T) {
	service, nc := setupTestService(t)
	defer nc.Close()

	// Get an existing room
	roomsResp, err := service.ListRooms(context.Background(), &pb.ListRoomsRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, roomsResp.Rooms)

	testRoom := roomsResp.Rooms[0]
	testUserID := uuid.New().String()

	// First join the room
	joinResp, err := service.JoinRoom(context.Background(), &pb.JoinRoomRequest{
		RoomId: testRoom.Id,
		UserId: testUserID,
	})
	require.NoError(t, err)
	require.True(t, joinResp.Success)

	// Now leave the room
	leaveResp, err := service.LeaveRoom(context.Background(), &pb.LeaveRoomRequest{
		RoomId: testRoom.Id,
		UserId: testUserID,
	})
	assert.NoError(t, err)
	assert.True(t, leaveResp.Success)
}
