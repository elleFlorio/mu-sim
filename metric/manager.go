package metric

import (
	"time"

	"github.com/influxdb/influxdb/client/v2"
)

type InfluxConfig struct {
	Address  string
	DBname   string
	Username string
	Password string
}

var (
	tags           map[string]string
	execFields     map[string]interface{}
	respTimeFields map[string]interface{}
	influx         client.Client
	config         InfluxConfig
	batch          client.BatchPoints
)

func Initialize(serviceName string, serviceWorkload string, serviceAddress string, influxConf InfluxConfig) error {
	var err error
	tags = map[string]string{
		"name":     serviceName,
		"workload": serviceWorkload,
		"address":  serviceAddress,
	}
	execFields = map[string]interface{}{
		"value": 0.0,
	}
	respTimeFields = map[string]interface{}{
		"value": 0.0,
	}
	config = influxConf

	influx, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     config.Address,
		Username: config.Username,
		Password: config.Password,
	})
	if err != nil {
		return err
	}

	batch, err = client.NewBatchPoints(client.BatchPointsConfig{
		Database:  config.DBname,
		Precision: "ms",
	})
	if err != nil {
		return err
	}

	return nil
}

func SendExecutionTime(execTime float64) error {
	execFields["value"] = execTime
	point, err := client.NewPoint("execution_time", tags, execFields, time.Now())
	if err != nil {
		return err
	}

	batch.AddPoint(point)
	influx.Write(batch)
	return nil
}

func SendResponseTime(respTime float64) error {
	respTimeFields["value"] = respTime
	point, err := client.NewPoint("response_time", tags, respTimeFields, time.Now())
	if err != nil {
		return err
	}

	batch.AddPoint(point)
	influx.Write(batch)
	return nil
}
