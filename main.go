package main

import (
	"github.com/HosanaUFRRJ2014/planets-api/model"
	"github.com/HosanaUFRRJ2014/planets-api/api"
)

func main() {

	const (
		// API configs
		host string = "127.0.0.1"
		port string = "5555"

		// Database configs
		dbHost string = "planetsapi-6dhqu.gcp.mongodb.net"
		dbUser string = "planetsapi"
		dbPassword string = "940bf335-e572-496e-af0e-10ec3b9e1cd3"
		databaseName string = "planetsapi"
		collectionName string = "planets"
	)

	client := model.MongoDBConnect(
		dbHost,
		dbUser,
		dbPassword,
		databaseName,
		collectionName,
	)
	defer model.MongoDBDisconnect(client)

	api.HandleRequests(host, port)
}
