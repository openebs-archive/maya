package snapshot

import (
	"fmt"
	"log"

	pb "github.com/openebs/maya/cmd/mayactl/app/command/cstor/OpenEBS"
	"github.com/spf13/cobra"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

type CmdSnapshotOptions struct {
	Name         string
	Volume       string
	Pool         string
	Server       string
	RemoteCstor  string
	RemoteUser   string
	RemotePass   string
	RemoteVolume string
	RemotePool   string
}

func NewCreateCmd() *cobra.Command {
	options := CmdSnapshotOptions{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			//options.Cmd = pb.Cmd_POOL
			//Validate(cmd)
			options.RunSnapshotCreate()
		},
	}
	cmd.Flags().StringVarP(&options.Name, "name", "", "", "unique snapshot name.")
	cmd.Flags().StringVarP(&options.Volume, "volume", "", "", "unique snapshot name.")
	cmd.Flags().StringVarP(&options.Pool, "pool", "", "", "unique snapshot name.")
	cmd.Flags().StringVarP(&options.Server, "server", "", "", "unique snapshot name.")
	return cmd
}

func (c *CmdSnapshotOptions) RunSnapshotCreate() error {
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
		Cmd:    pb.Cmd_SNAPSHOT,
		Name:   c.Name,
		Volume: c.Volume,
		Pool:   c.Pool,
	}
	out, err := client.Create(context.Background(), Request)
	fmt.Println(out.GetOutput())
	return nil
}
