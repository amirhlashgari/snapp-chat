package store

import (
	"testing"
	"time"

	pb "github.com/amirhlashgari/snapp-chat/proto"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestNATS(t *testing.T) *nats.Conn {
	nc, err := nats.Connect(nats.DefaultURL)
	require.NoError(t, err, "Failed to connect to NATS")
	return nc
}

func TestNewJetStreamStore(t *testing.T) {
	nc := setupTestNATS(t)
	defer nc.Close()

	store, err := NewJetStreamStore(nc)
	assert.NoError(t, err, "Should create JetStreamStore successfully")
	assert.NotNil(t, store, "Store should not be nil")
}

func TestSaveAndGetMessage(t *testing.T) {
	nc := setupTestNATS(t)
	defer nc.Close()

	store, err := NewJetStreamStore(nc)
	require.NoError(t, err)

	testMsg := &pb.Message{
		Id:        uuid.New().String(),
		RoomId:    "test-room",
		UserId:    "user1",
		Content:   "Hello, World!",
		Timestamp: time.Now().Unix(),
		Username:  "username",
	}
	err = store.SaveMessage(testMsg)
	assert.NoError(t, err, "Should save message successfully")

	messages, err := store.GetMessages("test-room", 1)
	assert.NoError(t, err, "Should retrieve messages successfully")
	assert.Equal(t, testMsg.Content, messages[0].Content, "Retrieved message should match saved message")
}

func TestSaveAndGetUser(t *testing.T) {
	nc := setupTestNATS(t)
	defer nc.Close()

	store, err := NewJetStreamStore(nc)
	require.NoError(t, err)

	testUser := &pb.User{
		Id:       "user1",
		Username: "testuser",
	}

	err = store.SaveUser(testUser)
	assert.NoError(t, err, "Should save user successfully")

	users, err := store.GetUsers()
	assert.NoError(t, err, "Should retrieve users successfully")
	assert.True(t, len(users) > 0, "Should have at least one user")
}

func TestSaveAndGetRoom(t *testing.T) {
	nc := setupTestNATS(t)
	defer nc.Close()

	store, err := NewJetStreamStore(nc)
	require.NoError(t, err)

	testRoom := &pb.ChatRoom{
		Id:   "room1",
		Name: "Test Room",
	}

	err = store.SaveRoom(testRoom)
	assert.NoError(t, err, "Should save room successfully")

	rooms, err := store.GetRooms()
	assert.NoError(t, err, "Should retrieve rooms successfully")
	assert.True(t, len(rooms) > 0, "Should have at least one room")
}
