package cli

import (
	"log"
	"strconv"

	"github.com/elleFlorio/mu-sim/Godeps/_workspace/src/github.com/codegangsta/cli"

	"github.com/elleFlorio/mu-sim/app"
	"github.com/elleFlorio/mu-sim/network"
)

func start(c *cli.Context) {
	if !c.Args().Present() {
		log.Fatalln("Cannot start service: service name is missing")
	}

	name := c.Args().First()
	etcdAddress := c.String("etcdserver")

	influxAddress := c.String("influxdb")
	influxDB := c.String("db-name")
	influxUser := c.String("db-user")
	influxPwd := c.String("db-pwd")

	var ip string
	if ip = c.String("ipaddress"); ip == "" {
		ip = network.GetHostIp()
	}

	var port string
	if port = c.String("port"); port == "" {
		p := network.GetPort()
		port = strconv.Itoa(p)
	}
	port = ":" + port

	workload := c.String("workload")
	destinations := c.StringSlice("destination")

	params := app.ServiceParams{
		etcdAddress,
		influxAddress,
		influxDB,
		influxUser,
		influxPwd,
		ip,
		port,
		name,
		workload,
		destinations,
	}

	app.StartService(params)
}
