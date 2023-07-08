# Bacalhau Operator

 This project aims to use Kubernetes as controlplane with [Bacalhau](https://docs.bacalhau.org/) as the orchestrator. This is a POC and not meant for production use.


### Prerequisites
* Kubernetes cluster- although any Kubernetes cluster(with recent version) would work, the light weight option would be to use [KCP](https://github.com/kcp-dev/kcp) as KCP doesn't have any orchestration components.
* Go 1.20+ installed 
* Bacalhau CLI installed

### Setup
* Clone the repo
* Run `kubectl create -f config/crd/bases/` to create the CRD
* Run `make run` to run the operator


### Usage
* Create a namespace called `Bacalhau`:
```bash
kubectl create namespace bacalhau
````

* Create a Job CR, a sample can be found at config/samples-
```bash
kubectl create -f config/samples/
````

* Check status of the job
```bash
kubectl -n bacalhau get job job-sample
```

* Grab the job ID from the status and check the status of the job in Bacalhau
```bash
bacalhau describe <job ID> | less
```

### TODO
* Add support for WASM
* Add support for resource requirements
* Accept input locations
* Reporting Bacalhau Job status back to Job CR. 


### Contributing
Please feel free to open issues for any bugs or feature requests. Pull requests are welcome too.