### Notice: worldPing has reached End-Of-Life (EOL) on April 1, 2021

Everything you love about worldPing is now available with the new synthetic monitoring feature in Grafana Cloud, plus you’ll get reduced complexity and all the benefits of Grafana Cloud. 

Take the first step to get started with synthetic monitoring by signing up for a Grafana Cloud account. [Get started for free.](https://go2.grafana.com/worldPing-EOL-grafana-cloud.html?pg=plugins-wp&plcmt=body-txt)

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
  * Download the src and dependencies (you need to have [Golang >= 1.13](https://golang.org/) [downloaded](https://golang.org/dl/) and [installed](https://golang.org/doc/install))
  ```
git clone https://github.com/raintank/raintank-probe
cd raintank-probe/cmd/raintank-probe
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


