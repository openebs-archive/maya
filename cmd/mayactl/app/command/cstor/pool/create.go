package pool

import (
	"fmt"
	"log"

	pb "github.com/openebs/maya/cmd/mayactl/app/command/cstor/OpenEBS"
	"github.com/spf13/cobra"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

type CmdPoolOptions struct {
	Name     string
	Interval string
	Server   string
}

func NewCreateCmd() *cobra.Command {
	options := CmdPoolOptions{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			options.RunPoolCreate()
		},
	}
	cmd.Flags().StringVarP(&options.Name, "name", "", "", "unique pool name.")
	return cmd
}

func (c *CmdPoolOptions) RunPoolCreate() error {
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
		Cmd:  pb.Cmd_POOL,
		Name: c.Name,
	}
	out, err := client.List(context.Background(), Request)
	fmt.Println(out.GetOutput())
	return nil
}
