# raintank-probe

Raintank probe package written in GO.

The raintank-probe provides the execution of periodic network performance tests including HTTP checks, DNS and Ping.
The results of each test are then transfered back to the Raintank API where they are processed and inserted into a timeseries database.

## To run your own private probe follow these steps.

1. Add the new probe via the raintank portal.
  * navigate to the probes page then click on the "New Probe" button at the top right of the screen.
  * enter a unique name for the probe and click the "add" button.
2. If you dont already have a Grafana.Net apiKey, [create one](https://grafana.net/profile).
3. Install the probe application - 2 options

  a.) Use Ubuntu Package. (Always the latest version)
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
server-url = wss://controller.raintank.io/
tsdb-url = https://tsdb.raintank.io/
api-key = <Your Grafana.net ApiKey>
public-checks = /etc/raintank/publicChecks.json
```
  * start the collector
  ```
  service raintank-probe start
  ```

  c.) Manual build of Raintank probe (Great for those wishing to test and contribute)
  * Download the src and dependencies (you need to have (Golang >= 1.5)[https://golang.org/]
Should be (downloaded)[https://golang.org/dl/] and (installed)[https://golang.org/doc/install])
  ```
go get github.com/raintank/raintank-probe
  ```
  * Create a config  with the probe name created in step 1 and the ApiKey created in step 2.
  ```
log-level = 2
name = <PROBE Name>
server-url = wss://controller.raintank.io/
tsdb-url = https://tsdb.raintank.io/
api-key = <Your Grafana.net ApiKey>
public-checks = /etc/raintank/publicChecks.json
```

  * Then start the app.
  ```
raintank-probe -config <path to your config>
  ```


