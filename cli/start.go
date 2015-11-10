package cli

import (
	"log"
	"strconv"

	"github.com/elleFlorio/testApp/Godeps/_workspace/src/github.com/codegangsta/cli"

	"github.com/elleFlorio/testApp/app"
	"github.com/elleFlorio/testApp/network"
)

func start(c *cli.Context) {
	if !c.Args().Present() {
		log.Fatalln("Cannot start service: service name is missing")
	}

	name := c.Args().First()
	etcdAddress := c.String("etcdserver")

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
	destinations := c.StringSlice("destinations")

	params := app.ServiceParams{
		etcdAddress,
		ip,
		port,
		name,
		workload,
		destinations,
	}

	app.StartService(params)
}
