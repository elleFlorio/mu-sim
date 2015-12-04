package cli

import (
	"log"
	"os"
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

	influxAddress := c.String("influxdb")
	influxDB := c.String("db-name")
	influxUser := os.Getenv("INFLUX_USER")
	influxPwd := os.Getenv("INFLUX_PWD")

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
