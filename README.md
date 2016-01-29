# MuSim - The Microservices Simulator#
This simple application let you simulate a graph of microservices communicating between each other. It is provided also like a Docker image you can start quickly with the right parameters. The application is designed to be a template that can be modified according to the needs of the user.
This is just a prototype, so use it at your own risk.

## How it Works ##
MuSim is a simulated microservice that simply keep busy the CPU performing a mathematical computation when a request arrive. The time for the computation is chosen according to an exponential distribution with a lambda that can vary according to the option set by the user. Thre resource usage is now limited to CPU, maybe in the future I will implement also something for the memory and the Disk operations.
Creating a set of MuSim is possible to simulate a graph of communicating microservices with dependencies among them. The execution time of every service, as well as its response time, can be monitored sending these metrics to an influxdb instance.

## Docs ##
This is the documentation on how to install MuSim and how to create a graph of microservices.

### Installation ###
You can go-get this repo with the command
[COMMAND]
if you want to modify the code and personalize the application. Otherwhise you can just pull the docker image:
[DOCKER IMAGE]

### Depencencies ###
MuSim requires a working instance of an etcd server [LINK] for service discovery.
Optionally you can setup an instance of influxdb to collect metrics (execution time and response time of services) about the status of the MuSim application.

### Usage ###
MuSim has a single "start" command that should be followed by the name of the service and some flags.
This is the list of required flags. the first column is the flag name with its abbreviation. The second column is the environment variable that can be set instead of the flag. The last column says if it is required for MuSim to run.
| flag | EnvVar | Description | Required |
| --- | --- | --- | --- |
| --etcdserver, -e | ETCD_ADDR | URL of etcd server | True |
| --ipaddress, -a | HostIP | address of the host | True if you run MuSim inside the Docker image, otherwise MuSim will automagically get the ip address|
| --port, -p | / | port of the service | True, but if not provided MuSim will automagically find a free port in the host |
| --workload, -w | / | Workload of the service. The value can be "none" (lambda = 0), "low" (lambda = 1s), "medium" (lambda = 5s), "high" (lambda = 10s) | False (default: "medium")|
| --destination, -d | / | Destination where to send the request once it has been completed. It can be used several times to set multiple destinations. The MuSim service waits for its destinations to responde before sending a response to the received request | False|
| --influxdb, -m | INFLUX_ADDR | URL of influxdb | False |
| --db-user, -dbu | INFLUX_USER | influxdb user username | False |
| --db-pwd, -dbp | INFLUX_PWD | influxdb user password | False |
| --db-name, -db | / | influxdb database name | False (default: "MuSimDB") |
