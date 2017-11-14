package snapshot

import (
	"fmt"
	"io"
	"log"

	pb "github.com/openebs/maya/cmd/mayactl/app/command/cstor/OpenEBS"
	"github.com/spf13/cobra"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

func NewSendCmd() *cobra.Command {
	options := CmdSnapshotOptions{}
	cmd := &cobra.Command{
		Use:   "send",
		Short: "",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			options.RunSnapshotSend()
		},
	}
	cmd.Flags().StringVarP(&options.Server, "server", "", "", "unique snapshot name.")
	cmd.Flags().StringVarP(&options.Name, "name", "", "", "unique snapshot name.")
	cmd.Flags().StringVarP(&options.Volume, "volume", "", "", "unique volume name.")
	cmd.Flags().StringVarP(&options.Pool, "pool", "", "", "unique pool name.")
	cmd.Flags().StringVarP(&options.RemoteCstor, "remotecstor", "", "", "IP address of remote cstor")
	cmd.Flags().StringVarP(&options.RemoteUser, "remoteuser", "", "", "username for remote cstor")
	cmd.Flags().StringVarP(&options.RemotePass, "remotepass", "", "", "password for remote cstor")
	cmd.Flags().StringVarP(&options.RemoteVolume, "remotevol", "", "", "volume for remote cstor")
	cmd.Flags().StringVarP(&options.RemotePool, "remotepool", "", "", "pool for remote cstor")
	return cmd
}

func (c *CmdSnapshotOptions) RunSnapshotSend() error {
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
		Cmd:         pb.Cmd_SNAPSHOT,
		Name:        c.Name,
		Volume:      c.Volume,
		Pool:        c.Pool,
		RemoteCstor: c.RemoteCstor,
		RemoteUser: c.RemoteUser,
		RemotePass: c.RemotePass,
		RemoteVolume: c.RemoteVolume,
		RemotePool: c.RemotePool,
	}

	stream, err := client.Send(context.Background(), Request)
	if err != nil {
		log.Fatalf("%v.Send(_) = _, %v", client, err)
	}
	for {
		out, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.Send(_) = _, %v", client, err)
		}
		fmt.Println(out.GetOutput())
	}

	//	out, err := client.Send(context.Background(), Request)
	//      fmt.Println(out.GetOutput())
	return nil
}
