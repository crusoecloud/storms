# Storage Management Service (StorMS)

This project hosts the implementation of the StorMS application, which is used to federate multiple storage clusters, agnostic of vendors.

## Design

StorMS is intended to be deployed as an application. The service is provides is management multiple clients to various storage backends (Lightbits, PureStorage, etc.) under a single name space. 

## Development/Running

If you update `.proto` files that that define StorMS, run `make proto` to generate Go files to reflect definition changes.

To run StorMS, first build the binary. Invokign the binary exposes its CLI, which will give instructions on how to start a StorMS instance.

```
make build # builds the StorMS binary into dist/
dist/storms -h # invoke the binary 
```

To start a StorMS instance, do

```
make build 
dist/storms serve
```

The `serve` command will start a running StorMS application using configuration specified in a user-provided (`--config`) or default file. 

For example:

```
# Given this working directory... 

storms (suppose, this is cwd)
├── dev
│   ├── clusters.yaml
│   └── storms.yaml
└── dist
    └── storms
```

```
~storms/ cat dev/storms.yaml

grpc_port: 9290
local_ip: 127.0.0.1
cluster_file: dev/clusters.yaml
```


```
~storms/ cat dev/clusters.yaml

clusters:
- vendor: "my-vendor"
  cluster_id: # provide a unique UUID here to identify the cluster
  vendor_config: # supports any kind of custom mapping
    api_key: "top-secret-api-key"
    endpoints: ["1.1.1.1", "2.2.2.2", "3.3.3.3"]
    cluster_tag: "some-tag"
```

```
~storms/ dist/storms --config dev/storms.yaml 

** StorMS should run at this point**
```
