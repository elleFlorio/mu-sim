package cli

import (
	"fmt"
	"os"

	"github.com/elleFlorio/testApp/Godeps/_workspace/src/github.com/codegangsta/cli"
)

func Run() {
	app := cli.NewApp()
	app.Name = "testApp"
	app.Usage = "Test application"

	app.Commands = []cli.Command{
		{
			Name:   "start",
			Usage:  "Start the a service",
			Action: start,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "etcdserver, e",
					Value: "http://localhost:4001",
					Usage: fmt.Sprintf("url of etcd server. Default is 'http://localhost:4001'"),
				},
				cli.StringFlag{
					Name:  "ipaddress, a",
					Value: "",
					Usage: fmt.Sprintf("Ip address of the host"),
				},
				cli.StringFlag{
					Name:  "port, p",
					Value: "",
					Usage: fmt.Sprintf("port of the service"),
				},
				cli.StringFlag{
					Name:  "workload, w",
					Value: "medium",
					Usage: fmt.Sprintf("workload (options: none, low, medium, heavy)"),
				},
				cli.StringSliceFlag{
					Name:  "destinations, d",
					Value: &cli.StringSlice{},
					Usage: fmt.Sprintf("destination of request messages. Can be used " +
						"several times to specify multiple destinations"),
				},
			},
		},
	}

	app.Run(os.Args)
}
