package app

import (
	"flag"
	"log"

	context "golang.org/x/net/context"
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	pb "github.com/openebs/cstor/OpenEBS"
	"google.golang.org/grpc"
)

func DestroyCmd() cli.Command {
	return cli.Command{
		Name: "destroy",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "server",
				Value: "localhost:50051",
			},
			cli.StringFlag{
				Name:  "cmd",
				Value: "snapshot",
			},
			cli.StringFlag{
				Name:  "name",
				Value: "",
			},
			cli.StringFlag{
				Name:  "pool",
				Value: "",
			},
			cli.StringFlag{
				Name:  "volume",
				Value: "",
			},
		},
		Action: func(c *cli.Context) {
			if err := destroy(c); err != nil {
				logrus.Fatalf("Error running controller command: %v.", err)
			}
		},
	}
}
func destroy(c *cli.Context) error {
	server := c.String("server")
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(server, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewCstorClient(conn)
	cmds := map[string]pb.Cmd{"pool": pb.Cmd_POOL, "volume": pb.Cmd_VOLUME, "snapshot": pb.Cmd_SNAPSHOT}
	Request := &pb.Request{
		Cmd: cmds[c.String("cmd")],
		Name: c.String("name"),
		Pool: c.String("pool"),
		Volume: c.String("volume"),
	}
	client.Destroy(context.Background(), Request)
	return nil
}
