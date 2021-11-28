# Rest-API
It's a simple REST API application written in Golang using GORM and PostgreSQL. To run you need [docker](https://www.docker.com/) to be installed  

## How to run
Build and run
```
make build && make run
```

If you run for the first time, make migration-up. Before it make sure your PostgreSQL server is stopped.
```
make migration-up
```
you can also clear database by using following command.
```
make migration-down
```
stop all containers
```
make stop
```
## In addition
run tests
```
make test
```
golangci-lint
```
make lint
```