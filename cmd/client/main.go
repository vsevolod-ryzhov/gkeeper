package main

import (
	//"context"
	"fmt"
	"gkeeper/internal/config"
	"os"

	//pb "gkeeper/api/proto"

	//"google.golang.org/grpc"
	//"google.golang.org/grpc/credentials/insecure"
	//"google.golang.org/protobuf/proto"

	"gkeeper/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	config.ParseFlags()
	//ctx := context.Background()
	//
	//conn, err := grpc.NewClient(config.Options.AppPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	//if err != nil {
	//	panic(err)
	//}
	//defer conn.Close()
	//
	//client := pb.NewGKeeperClient(conn)
	//
	//request, reqErr := client.Login(ctx, pb.LoginRequest_builder{
	//	Email:    proto.String("test@test.com"),
	//	Password: proto.String("12345qwert"),
	//}.Build())
	//
	//if reqErr != nil {
	//	panic(reqErr)
	//}
	//
	//fmt.Println(request.GetResult())

	p := tea.NewProgram(
		tui.NewMainModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
