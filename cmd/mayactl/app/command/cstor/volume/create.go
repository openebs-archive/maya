package volume

import (

	"log"
	"fmt"

	context "golang.org/x/net/context"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	pb "github.com/openebs/maya/cmd/mayactl/app/command/cstor/OpenEBS"
)

type CmdVolumeOptions struct {
        Name string
}

func NewCreateCmd() *cobra.Command {
	options := CmdVolumeOptions{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			options.RunVolumeCreate()
		},
	}
	cmd.Flags().StringVarP(&options.Name, "name", "", "", "unique volume name.")
	return cmd
}

func (c *CmdVolumeOptions) RunVolumeCreate() error {
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
		Cmd: pb.Cmd_VOLUME,
		Name: c.Name,
	}
	out, err := client.List(context.Background(), Request)
        fmt.Println(out.GetOutput())
	return nil
}
