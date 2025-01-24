package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/amirhlashgari/snapp-chat/internal/client"
	pb "github.com/amirhlashgari/snapp-chat/proto"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const menu = `
Chat Application Menu:
1. List Users
2. List Rooms
3. Join Room
4. Leave Room
5. Exit
Enter your choice: `

func main() {
	username := flag.String("user", "", "Username for chat")
	natsURL := flag.String("nats", nats.DefaultURL, "NATS server URL")
	serviceAddr := flag.String("service", "localhost:50051", "Chat service address")
	flag.Parse()

	if *username == "" {
		log.Fatal("Username is required")
	}

	nc, err := nats.Connect(*natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}

	conn, err := grpc.NewClient(*serviceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to service: %v", err)
	}
	defer conn.Close()

	service := pb.NewChatServiceClient(conn)

	userID := uuid.New().String()
	client, err := client.NewClient(userID, *username, nc, service)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		client.Close()
		os.Exit(0)
	}()

	go receiveMessages(client)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(menu)
		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())
		switch choice {
		case "1":
			listUsers(client, scanner)
		case "2":
			listRooms(client, scanner)
		case "3":
			joinRoom(client, scanner)
		case "4":
			leaveRoom(client)
		case "5":
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

func listUsers(client *client.Client, scanner *bufio.Scanner) {
	fmt.Print("Enter filter (or press Enter for all): ")
	scanner.Scan()
	filter := strings.TrimSpace(scanner.Text())

	users, err := client.ListUsers(filter)
	if err != nil {
		fmt.Printf("Error listing users: %v\n", err)
		return
	}

	fmt.Println("\nUsers:")
	for i, user := range users {
		fmt.Printf("%d. %s (%s)\n", i+1, user.Username, user.Status)
	}
}

func listRooms(client *client.Client, scanner *bufio.Scanner) {
	fmt.Print("Enter filter (or press Enter for all): ")
	scanner.Scan()
	filter := strings.TrimSpace(scanner.Text())

	rooms, err := client.ListRooms(filter)
	if err != nil {
		fmt.Printf("Error listing rooms: %v\n", err)
		return
	}

	fmt.Println("\nRooms:")
	for i, room := range rooms {
		fmt.Printf("%d. %s (%d members)\n", i+1, room.Name, len(room.Members))
	}
}

func joinRoom(client *client.Client, scanner *bufio.Scanner) {
	rooms, err := client.ListRooms("")
	if err != nil {
		fmt.Printf("Error listing rooms: %v\n", err)
		return
	}

	fmt.Println("\nAvailable Rooms:")
	for i, room := range rooms {
		fmt.Printf("%d. %s\n", i+1, room.Name)
	}

	fmt.Print("Enter room number to join: ")
	scanner.Scan()
	choice := strings.TrimSpace(scanner.Text())

	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(rooms) {
		fmt.Println("Invalid room number")
		return
	}

	room := rooms[idx-1]
	if err := client.JoinRoom(room.Id); err != nil {
		fmt.Printf("Error joining room: %v\n", err)
		return
	}

	fmt.Printf("Joined room: %s\n", room.Name)
	chatMode(client, scanner)
}

func leaveRoom(client *client.Client) {
	if err := client.LeaveRoom(""); err != nil {
		fmt.Printf("Error leaving room: %v\n", err)
		return
	}
	fmt.Println("Left room")
}

func chatMode(client *client.Client, scanner *bufio.Scanner) {
	fmt.Println("\nChat Mode (type /exit to leave):")
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "/exit" {
			return
		}

		if err := client.SendMessage(input); err != nil {
			fmt.Printf("Error sending message: %v\n", err)
			return
		}
	}
}

func receiveMessages(client *client.Client) {
	for msg := range client.MessageChannel() {
		unixTimeUTC := time.Unix(msg.Timestamp, 0)
		unitTimeInRFC3339 := unixTimeUTC.Format(time.RFC3339)

		fmt.Printf("\n[%s] - [%s]: %s \n", msg.Username, unitTimeInRFC3339, msg.Content)
	}
}
