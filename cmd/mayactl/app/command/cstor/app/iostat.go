package app

import (
	"flag"
	"fmt"
	"log"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	pb "github.com/openebs/cstor/OpenEBS"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

func IOStatsCmd() cli.Command {
	return cli.Command{
		Name: "iostats",
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
			if err := iostats(c); err != nil {
				logrus.Fatalf("Error running controller command: %v.", err)
			}
		},
	}
}
func iostats(c *cli.Context) error {
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
	}
	fmt.Println(cmds[c.String("cmd")])
	out, err := client.IOStats(context.Background(), Request)
	fmt.Println(out.GetOutput())
	return nil
}
