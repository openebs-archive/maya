package pool

import (
	"fmt"
	"io"
	"log"

	pb "github.com/openebs/maya/cmd/mayactl/app/command/cstor/OpenEBS"
	"github.com/spf13/cobra"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

func NewIOStatsCmd() *cobra.Command {
	options := CmdPoolOptions{}
	cmd := &cobra.Command{
		Use:   "iostats",
		Short: "",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			//options.Cmd = pb.Cmd_POOL
			//Validate(cmd)
			options.RunPoolIOStats()
		},
	}
	cmd.Flags().StringVarP(&options.Name, "name", "", "", "unique pool name.")
	cmd.Flags().StringVarP(&options.Interval, "interval", "", "", "time gap between stats in seconds")
	cmd.Flags().StringVarP(&options.Server, "server", "", "", "Server")
	return cmd
}

func (c *CmdPoolOptions) RunPoolIOStats() error {
	server := c.Server
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(server, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewCstorClient(conn)
	Request := &pb.Request{
		Cmd:      pb.Cmd_POOL,
		Name:     c.Name,
		Interval: c.Interval,
	}
	stream, err := client.IOStats(context.Background(), Request)
	if err != nil {
		log.Fatalf("%v.IOStats(_) = _, %v", client, err)
	}
	for {
		out, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.IOStats(_) = _, %v", client, err)
		}
		fmt.Println(out.GetOutput())
	}
	//out, err := client.IOStats(context.Background(), Request)
	//fmt.Println(out.GetOutput())
	return nil
}
