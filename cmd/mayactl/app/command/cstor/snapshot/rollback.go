package snapshot

import (

	"log"
	"fmt"

	context "golang.org/x/net/context"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	pb "github.com/openebs/maya/cmd/mayactl/app/command/cstor/OpenEBS"
)

func NewRollbackCmd() *cobra.Command {
	options := CmdSnapshotOptions{}
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			//Validate(cmd)
			options.RunSnapshotRollback()
		},
	}
	cmd.Flags().StringVarP(&options.Name, "name", "", "", "unique snapshot name.")
	return cmd
}

func (c *CmdSnapshotOptions) RunSnapshotRollback() error {
	server := "localhost:50051"
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(server, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewCstorClient(conn)
	Request := &pb.Request{
		Cmd: pb.Cmd_SNAPSHOT,
		Name: c.Name,
		Volume: c.Volume,
		Pool: c.Pool,
	}
	out, err := client.Rollback(context.Background(), Request)
        fmt.Println(out.GetOutput())
	return nil
}
