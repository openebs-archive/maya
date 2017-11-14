package volume

import (

	"log"
	"fmt"

	context "golang.org/x/net/context"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	pb "github.com/openebs/maya/cmd/mayactl/app/command/cstor/OpenEBS"
)

func NewListCmd() *cobra.Command {
	options := CmdVolumeOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			options.RunVolumeList()
		},
	}
	cmd.Flags().StringVarP(&options.Name, "name", "", "", "unique volume name.")
	return cmd
}

func (c *CmdVolumeOptions) RunVolumeList() error {
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
		Cmd: pb.Cmd_VOLUME,
	}
	out, err := client.List(context.Background(), Request)
        fmt.Println(out.GetOutput())
	return nil
}
