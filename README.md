# MuSim - The Microservice Simulator#
This simple application let you simulate a graph of microservices communicating between each other. The application is designed to be a template that can be modified according to the needs of the user.
This is just a prototype, so use it at your own risk.

## How it Works ##
MuSim is a simulated microservice that simply keep busy the CPU performing a mathematical computation when a request arrive. The time for the computation is chosen according to an exponential distribution with a lambda that can vary according to the option set by the user. Thre resource usage is now limited to CPU, maybe in the future I will implement also something for the memory and the Disk operations.
Creating a set of MuSim is possible to simulate a graph of communicating microservices with dependencies among them. The execution time of every service, as well as its response time, can be monitored sending these metrics to an influxdb instance.

## Docs ##
This is the documentation on how to install MuSim and how to use it in order to create a graph of microservices.

### Installation ###
You can go-get this repo with the command

`go get github.com/elleFlorio/mu-sim`

if you want to modify the code and personalize the application. Otherwhise you can just pull the docker image:

`docker pull elleflorio/mu-sim`

### Dependencies ###
MuSim requires a working instance of an [etcd](https://github.com/coreos/etcd) server for service discovery.
Optionally you can setup an instance of [influxdb](https://github.com/influxdata/influxdb) to collect metrics (execution time and response time of services) about the status of the MuSim application.

### Usage ###

##### How to start a MuSim #####
MuSim has a single "start" command that should be followed by the name of the service and some flags.
This is the list of required flags. the first column is the flag name with its abbreviation. The second column is the environment variable that can be set instead of the flag. The last column says if it is required for MuSim to run.

| Flag | EnvVar | Description | Required |
| --- | --- | --- | --- |
| etcdserver, e | ETCD_ADDR | URL of etcd server | True |
| ipaddress, a | HostIP | address of the host | True if you run MuSim inside the Docker container, otherwise MuSim will automagically get the ip address |
| port, p | / | port of the service | True, but if not provided MuSim will automagically find a free port in the host |
| workload, w | / | Workload of the service. The value can be "none" (lambda=0s), "low" (lambda=1s), "medium" (lambda=5s), "heavy" (lambda=10s) | False (default: "medium") |
| destination, d | / | Destination where to send the request once it has been completed. It can be used several times to set multiple destinations. The MuSim service waits for ALL its destinations to respond before sending a response to the received request | False |
| influxdb, m | INFLUX_ADDR | URL of influxdb | False |
| db-user, dbu | INFLUX_USER | influxdb user username | False |
| db-pwd, dbp | INFLUX_PWD | influxdb user password | False |
| db-name, db | / | influxdb database name | False (default: "MuSimDB") |

##### How to send requests to MuSim ####
The requests to the MuSim should be sent as http POST request with a JSON content/type formatted in this way:

`{"sender":"","body":"do", "args":""}`

You can also specify a destination directly inside the request. Suppose you want to send a request to the MuSim pippo (listening at `http://localhost:8080`) and then pippo should send that request to the MuSim topolino, the following command does the trick:

`curl -H "Content-Type: application/json" -X POST -d '{"sender":"","body":"do", "args":""}' http://localhost:8080/message?service=topolino`

##### Load balancing #####
MuSim automatically load balance the requests to its destinations selecting randomly a target in the set of the instances of the destination. Let's clarify this with an example:
suppose the MuSim pippo has the MuSim topolino as destination, and MuSim topolino has 3 active instances (i.e. there are 3 MuSim started with name "topolino"). The MuSim pippo asks to the etcd server the active instances of MuSim topolino, then chose randomly (uniform distribution) one of the instances as the destination of the request.

##### Scaling #####
MuSim register itself to the etcd server when it starts and then run a "keepAlive" function to notify etcd that it is still there up and running. This means that you can start and stop MuSim instances without worries. When a MuSim is stopped it follows this shut down steps:
- stop the "keepAlive" function (so it won't receive requests anymore)
- compute the pending requests and send the requests to the destinations
- wait for the response of the destinations
- respond to the requests
- shut down
This procedure is not 100% robust, and may lead to a deadlock if there is a failure in one of the destination that is computing a response. However, it is enough to ensure the scaling (also automatic) of the application.

##### Fault tolerance #####
MuSim is not fault tolerant by now and requests may be lost due to failure of MuSim instances. Maybe someday I will implement a mechanism to handle failures, but now it is up to you.

### Examples ###
* Start a single service named "pippo" using Env Vars with default parameters and no destinations:

`mu-sim start pippo`

* Start a single service named "pippo" using Env Vars with workload "low" and destinations topolino and paperino:

`mu-sim start pippo -w low -d topolino -d paperino`

* Start a single service named "pippo" using Env Vars using the docker image and specifying ip and port

`docker run -e ETCD_ADDR -e HostIP -e INFLUX_USER -e INFLUX_PWD -e INFLUX_ADDR -p 50100:50100 --name pippo elleflorio/musim start pippo -p 50100`

* Create the following graph of MuSims

![MuSim graph](https://github.com/elleFlorio/mu-sim/blob/master/mu-sim_graph.png)

The endpoint does not execute any work, it simply acts as the entrypoint of the application and balance the load among the destinations. Destinations are not specified, because want to send a request to service1a or service1b but not both of them, so the destination will be specified inside the requests. Service1a and service1b execute a medium workload; service2a, service2b and service2c execute a heavy workload; the database execute a low workload.

Let's create all the MuSims:
  - `endpoint: mu-sim start endpoint -w none -p 8080`
  - `service1a: mu-sim start service1a -d service2a -d service2b`
  - `service1b: mu-sim start service1b -d service2b -d service2c`
  - `service2a: mu-sim start service2a -w heavy -d database`
  - `service2b: mu-sim start service2b -w heavy -d database`
  - `service2c: mu-sim start service2c -w heavy -d database`
  - `database: mu-sim start database -w low`

Now you can send the requests to the endpoint both to service1a:

`curl -H "Content-Type: application/json" -X POST -d '{"sender":"","body":"do", "args":""}' http://localhost:8080/message?service=service1a`

or to service1b:

`curl -H "Content-Type: application/json" -X POST -d '{"sender":"","body":"do", "args":""}' http://localhost:8080/message?service=service1b`

## That's all ##
Feel free to contribute or send me feedbacks/comments/requests

That's all, enjoy! :-)
