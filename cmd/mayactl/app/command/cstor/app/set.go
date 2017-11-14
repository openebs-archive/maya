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

func SetCmd() cli.Command {
	return cli.Command{
		Name: "set",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "server",
				Value: "localhost:50051",
			},
			cli.StringFlag{
				Name:  "cmd",
				Value: "volume",
			},
		},
		Action: func(c *cli.Context) {
			if err := set(c); err != nil {
				logrus.Fatalf("Error running controller command: %v.", err)
			}
		},
	}
}
func set(c *cli.Context) error {
	server := c.String("server")
	flag.Parse()
	var opts []grpc.DialOption
	conn, err := grpc.Dial(server, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewCstorClient(conn)
	cmds := map[string]pb.Cmd{"pool": pb.Cmd_POOL, "volume": pb.Cmd_VOLUME, "snapshot": pb.Cmd_SNAPSHOT}
	Request := &pb.Request{
		Cmd: cmds[c.String("cmd")],
	}
	client.Set(context.Background(), Request)
	return nil
}
