## Demo

This is a demo that will run a small kubernetes cluster on your macbook for influx2.0 to run.  CVP is assuming reachable on your network.

Requirements are as follows
- kubectl
- kind

Both are pretty easy installs if you do not have them
```
brew install kind
brew install kubernetes-cli
```


### Credentials
Influx:
 <br>
 <br>arista/arista123!

#### Create the cluster localy if needed.

```
kind create cluster --config=demo/kind/config.yaml
```

#### Create the monitoring namespace

```
 kubectl create namespace monitoring
```

#### Apply the influx manifests

```
kubectl apply -f demo/influxmanifests/.
```

As long as it looks similar to this influx would have started.
```
kubectl get pods -n monitoring
NAME                   READY   STATUS      RESTARTS   AGE
influxdb-0             1/1     Running     0          72s
influxdb-setup-pmjwj   0/1     Error       0          72s
influxdb-setup-w8h64   0/1     Completed   0          41s
```

Do not worry about the setup batch pod as its only job is to talk to the influx api and create a bucket.  It will run until completion.  So as long as one Completed then it should work.

#### Proxy your traffic to the kind host.
```
kubectl port-forward service/influxdb 8086:8086 -n monitoring --address 0.0.0.0 &
```

You should be able to login to http://127.0.0.1:8086/signin<br>
username: arista password; arista123!

### Edit the config.yaml file with the correct info for the demo
<br>
my config.yaml file looks as follows

```
cvp_server: "10.90.226.175:8443
influxurl: "http://127.0.0.1:8086"
influxorg: "InfluxData"
influxbucket: "kubernetes"
cvptoken: "cvptokenhere"
influxtoken: "secret-token"
measurement: "interfaces"
path: "/interfaces/interface/state/counters"
origin: openconfig
streammode: on_change
```

The cvptoken is rather large but the rest should work.

## Run the code
```
cd bin
./cvp-to-influx-darwin -config ../config.yaml
```
Do keep in mind for every transaction this will do it will log inside of influx as well as the logging on the binary.