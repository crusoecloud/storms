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
cd dist
storms serve
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
~storms/dist storms app start --config dev/storms.yaml 

** StorMS should run at this point**
```

## Supported Vendors
### Lightbits
```
clusters:
  - vendor: "lightbits"
    cluster_id: <uuid>
    vendor_config:
      auth_token: <lightbits-jwt-token>
      addr_strs: 
        - `<host-ip>:<port>` # example: 172.2.1.12:443
          `<host-ip>:<port>` 
      project_name: <[dev,staging,prod]>
      replication_factor: <[2,3]>
```

### PureStorage
```
clusters:
  - vendor: "purestorage"
    cluster_id: <uuid>
    vendor_config:
      ## TODO ##
```

### Krusoe (this is a mock backend for development purposes)
```
clusters:
  - vendor: "krusoe"
    cluster_id: <uuid>
    vendor_config:
      api_key: krusoe # this a hard-coded password 
```

## Multi-cluster, multi-vendor example

In this example, we will configure StorMS to manage 4 clusters: x2 Lightbits cluster, x1 PureStorage cluster, and x1 Krusoe cluster.

You can have the following StorMS application configuration to StorMS serve using the endpoint `127.0.0.1:9290` and target a cluster configuration file at the path specified in `cluster_file`

```
# /storms/storms.yaml

grpc_port: 9290
local_ip: 127.0.0.1
cluster_file: /storms/clusters.yaml
```

Then, the cluster configuration file may contain

```
# /storms/clusters.yaml

clusters:
  - vendor: lightbits
    cluster_id: 9a3f1621-92e5-4d2c-8892-0d67e74e90d2
    vendor_config:
      auth_token: SECRET_JWT_TOKEN
      addr_strs: ['1.1.1.1:443', '2.2.2.2:443]
      project_name: staging
      replication_factor: 2
  - vendor: lightbits
    cluster_id: a62495e1-ed98-4229-af42-9c9b0935d508
    vendor_config:
      auth_token: SECRET_JWT_TOKEN
      addr_strs: 
        - '3.3.3.3:443'
        - '4.4.4.4:443'
      project_name: staging
      replication_factor: 3
  - vendor: purestorage
    cluster_id: 3b73e389-a1a8-4520-9eb6-39f47c449f9e
    vendor_config:
      ## TODO ##
  - vendor: krusoe
    cluster_id: 00b35e88-06b1-4ea4-89ad-49407217454f
    vendor_config:
      api_key: krusoe
```