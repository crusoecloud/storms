<!-- # Go Template

This project is used as a template for Crusoe's Go projects. It contains the common CI/CD files and directory structure that we use in our projects.

## Usage

### Setup
First, you should rename the module by changing the first line in `go.mod`. The same change should be made in the go files included here if you decide to keep them. Look for instances of `TODO(template)` in this repo for places where you might need to change something.

### Go executable
If you're creating an executable, rename the subdir under `cmd` to the name of your executable and use the `main.go` file there as the entrypoint. It should be a thin entrypoint and depend on packages in the `internal` dir for the actual logic. You'll also need to rename the project name and executable name in `Makefile`.

### Go library
If you're creating a library, delete the `cmd` dir. Any packages you want to be exported from the library (i.e. that you want to depend on in other repos) should be in the top level dir of the repo. Any that you don't want to be exported should go in the `internal` dir.

### Go versioning
This template uses Go 1.23, but note that some Crusoe repos might be on older/newer versions. If you need to update the go version, do so in `go.mod` and the `CI_IMAGE` variable in `.gitlab-ci.yml`. 

### Other files
Update the Makefile as you see fit, but make sure to have a `ci` directive that can be used by our Gitlab runner to verify commits. Please see the `.gitlab-ci.yml` file for the base CI setup. You can also update that file if you need to perform other types of operations during CI.

The `.golangci.yml` file should be left untouched unless there is a good reason to change the settings for or delete one of the linters. When submitting a MR that does remove support for a linter, please provide the reasoning in an inline comment in that file. -->

# Storage Management Service (StorMS)

This project hosts the implementation of the StorMS application, which is used to federate multiple storage clusters, agnostic of vendors.

## Design

StorMS is intended to be deployed as an application. The service is provides is management multiple clients to various storage backends (Lightbits, PureStorage, etc.) under a single name space. 

## Development/Running

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