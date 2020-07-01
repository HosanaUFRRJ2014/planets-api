# planets-api
Star Wars Planets Golang API

[Trello for this board (In Portuguese)](https://trello.com/b/WWVcIcWl/api-planetas)

## Requirements:
    
- [Go](https://golang.org/) ==(v1.10.x)
- [Dep](https://golang.github.io/dep/) >=([v0.5.4](https://github.com/golang/dep/releases))
- [Docker](https://docs.docker.com/get-docker/)
- [Docker-Compose](https://docs.docker.com/compose/install/)

## Instalation:
### You *might* need to install Golang dependencies
    dep ensure

### Download Mongo image
    docker pull mongo

## Usage:

### Run docker container for mongo
    docker-compose docker-compose.yml up

### Run the API:

Directly from source

    go run main.go 

or

... build and run the file

    go build main.go
    ./main
or 

...run the build, if you are on Linux/Ubuntu x64

    ./main

Access `localhost:5555/planets/api` and check the list what this API can do.

## Extras:
### Configuring API's port:
The default API port is 5555, but you may want to run it in a diferent one.

For this, run:

    go main.go -port {PORT}

or...

    ./main -port {PORT}

Check `./main -help` for more config options.
Note: If you want to change db's configs, you will also need to edit `docker-compose.yml`.
