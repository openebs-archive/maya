package snapshot

import (

	"log"
	"fmt"

	context "golang.org/x/net/context"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	pb "github.com/openebs/maya/cmd/mayactl/app/command/cstor/OpenEBS"
)

func NewDestroyCmd() *cobra.Command {
	options := CmdSnapshotOptions{}
	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			//options.Cmd = pb.Cmd_POOL
			//Validate(cmd)
			options.RunSnapshotDestroy()
		},
	}
	cmd.Flags().StringVarP(&options.Name, "name", "", "", "unique snapshot name.")
	return cmd
}

func (c *CmdSnapshotOptions) RunSnapshotDestroy() error {
	server := "localhost:50051"
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(server, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewCstorClient(conn)
	fmt.Println(c.Name)
	Request := &pb.Request{
		Cmd: pb.Cmd_SNAPSHOT,
		Name: c.Name,
		Volume: c.Volume,
		Pool: c.Pool,
	}
	out, err := client.Destroy(context.Background(), Request)
        fmt.Println(out.GetOutput())
	return nil
}
