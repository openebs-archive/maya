package main

import (
	"fmt"
	"log"

	"github.com/openebs/maya/cmd/cstor-volume-grpc/api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func creatVolumeSnapshot() {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf(":%d", api.VolumeGrpcListenPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := api.NewRunCommandClient(conn)
	response, err := c.RunVolumeCommand(context.Background(),
		&api.VolumeCommand{
			Command:  api.CmdSnapCreate,
			Volume:   "testvol1",
			Snapname: "testsnap1",
			Status:   "requesting",
		})

	if err != nil {
		log.Fatalf("Error when calling RunVolumeCommand: %s", err)
	}
	log.Printf("Response from server: %s, %s, %s, %s",
		response.Command, response.Volume, response.Snapname, response.Status)
}

func main() {
	creatVolumeSnapshot()
}
