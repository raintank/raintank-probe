# raintank-probe

Raintank probe package written in GO.

The raintank-probe provides the execution of periodic network performance tests including HTTP checks, DNS and Ping.
The results of each test are then transfered back to the Raintank API where they are processed and inserted into a timeseries database.

## To run your own private probe follow these steps.

1. Add the new probe via the raintank portal.
  * navigate to the probes page then click on the "New Probe" button at the top right of the screen.
  * enter a unique name for the probe and click the "add" button.
2. If you dont already have a Grafana.Net apiKey, [create one](https://grafana.net/profile).
3. Install the probe application - 4 options

  a.) Use Deb Package. Available for Ubuntu 14.04, Ubuntu 16.04, Debian Jessie
  * add PackageCloud to repo.
  ```
  curl -s https://packagecloud.io/install/repositories/raintank/raintank/script.deb.sh | sudo bash
  ```
  * Install raintank-probe package
  ```
  apt-get install raintank-probe
  ```
  * Edit the configuration file in /etc/raintank/probe.ini using the Probe name from step 1 and apiKey from step2
  ```
log-level = 2
name = <PROBE Name>
server-url = wss://worldping-api.raintank.io/
tsdb-url = https://tsdb-gw.raintank.io/
api-key = <Your Grafana.net ApiKey>
```
  * start the collector
  ```
  service raintank-probe start
  ```

  b.) Use RPM Package. Avalailable for Centos 6 and Centos 7 (and compatilble distrobutions.)
  * add PackageCloud to repo.
  ```
  curl -s https://packagecloud.io/install/repositories/raintank/raintank/script.rpm.sh | sudo bash
  ```
  * Install raintank-probe package
  ```
  yum install raintank-probe
  ```
  * Edit the configuration file in /etc/raintank/probe.ini using the Probe name from step 1 and apiKey from step2
  ```
log-level = 2
name = <PROBE Name>
server-url = wss://worldping-api.raintank.io/
tsdb-url = https://tsdb-gw.raintank.io/
api-key = <Your Grafana.net ApiKey>
```
  * start the collector
  ```
  service raintank-probe start
  ```

  c.) Use the Docker image.
  * launch the container with the below command, inserting the probe name from step1 and the apiKey from step2

  ```
  docker run -e RTPROBE_API_KEY=<Your Grafana.net ApiKey> -e RTPROBE_NAME=<PROBE name> raintank/raintank-probe 
  ```

  d.) Manual build of Raintank probe (Great for those wishing to test and contribute)
  * Download the src and dependencies (you need to have [Golang >= 1.9](https://golang.org/) [downloaded](https://golang.org/dl/) and [installed](https://golang.org/doc/install))
  ```
go get -d github.com/raintank/raintank-probe
cd $GOPATH/github.com/raintank/raintank-probe
go install -ldflags "-X main.GitHash=$(git describe --long --always)" 
  ```
  * Create a config  with the probe name created in step 1 and the ApiKey created in step 2.
  ```
log-level = 2
name = <PROBE Name>
server-url = wss://worldping-api.raintank.io/
tsdb-url = https://tsdb-gw.raintank.io/
api-key = <Your Grafana.net ApiKey>
```

  * Then start the app.
  ```
raintank-probe -config <path to your config>
  ```

## Note on Private Probe ICMP Health Checks
By default the private probe will determine its own network health by pinging comma separated list of sites google.com,youtube.com,facebook.com,twitter.com,wikipedia.com
If your private network does not allow ICMP traffic to external sites, you will need to modify probe.ini found here
```
 /etc/raintank/probe.ini 
```
Example probe.ini from docker container
```
log-level = 2
name = demo1
server-url = wss://worldping-api.raintank.io/
tsdb-url = https://tsdb-gw.raintank.io/
api-key = changeme
health-hosts = icmp-reachable-host.com
```
after modifying the health-hosts to an ICMP reachable host, restart. Then other checks you have assigned to the probe will begin to execute. 
