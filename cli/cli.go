package cli

import (
	"fmt"
	"os"

	"github.com/elleFlorio/mu-sim/Godeps/_workspace/src/github.com/codegangsta/cli"
)

func Run() {
	app := cli.NewApp()
	app.Name = "mu-sim"
	app.Usage = `Microservices simulator.
	This simple application allows to simulate a graph of microservices communicating between each other.
	The application is designed to be a template that can be modified according to needs of the user.
	This is just a prototype, so use it at your own risk.`
	app.Version = "1.0.0"
	app.Author = "Luca Florio (Github: elleFlorio)"
	app.Email = "elle.florio@gmail.com"

	app.Commands = []cli.Command{
		{
			Name:   "start",
			Usage:  "Start the a service",
			Action: start,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "etcdserver, e",
					Usage:  fmt.Sprintf("url of etcd server"),
					EnvVar: "ETCD_ADDR",
				},
				cli.StringFlag{
					Name:   "ipaddress, a",
					Value:  "",
					Usage:  fmt.Sprintf("Ip address of the host"),
					EnvVar: "HostIP",
				},
				cli.StringFlag{
					Name:   "influxdb, m",
					Usage:  fmt.Sprintf("url of influxdb"),
					EnvVar: "INFLUX_ADDR",
				},
				cli.StringFlag{
					Name:   "db-user, dbu",
					Usage:  fmt.Sprintf("influxdb user username"),
					EnvVar: "INFLUX_USER",
				},
				cli.StringFlag{
					Name:   "db-pwd, dbp",
					Usage:  fmt.Sprintf("influxdb user password"),
					EnvVar: "INFLUX_PWD",
				},
				cli.StringFlag{
					Name:  "db-name, db",
					Value: "muSimDB",
					Usage: fmt.Sprintf("influxdb database name. Default is 'testAppDB'"),
				},
				cli.StringFlag{
					Name:  "port, p",
					Value: "",
					Usage: fmt.Sprintf("port of the service"),
				},
				cli.StringFlag{
					Name:  "workload, w",
					Value: "medium",
					Usage: fmt.Sprintf("workload (options: none, low, medium, heavy). Default is 'medium'"),
				},
				cli.StringSliceFlag{
					Name:  "destination, d",
					Value: &cli.StringSlice{},
					Usage: fmt.Sprintf("destination of request messages. Can be used " +
						"several times to specify multiple destinations"),
				},
			},
		},
	}

	app.Run(os.Args)
}
