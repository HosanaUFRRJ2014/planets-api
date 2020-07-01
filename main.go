package main

import (
	"flag"
	"github.com/HosanaUFRRJ2014/planets-api/model"
	"github.com/HosanaUFRRJ2014/planets-api/api"
)

type Config struct {
	name string
	defaultValue string
	usage string
	valuePtr *string
}


func parseArgs(configs map[string]*Config) {
	for name, mapping := range configs {
		var valuePtr = flag.String(
			name,
			mapping.defaultValue,
			mapping.usage,
		)

		mapping.valuePtr = valuePtr
	}
}


func main() {
	var configs map[string]*Config
	configs = map[string]*Config{
		"host": &Config{
			name:"host",
			defaultValue:"127.0.0.1",
			usage:"Host which this API will be running.",
		},
		"port": &Config{
			name:"port",
			defaultValue:"5555",
			usage:"Port which this API will be running.",
		},
		"db_host": &Config{
			name:"db_host",
			defaultValue:"localhost",
			usage:"Host where this API's database will be running.",
		},
		"db_port": &Config{
			name:"db_port",
			defaultValue:"27017",
			usage:"Port where this API's database will be running.",
		},
		"db_user": &Config{
			name:"db_user",
			defaultValue:"planetsapi",
			usage:"User of this API's database.",
		},
		"database_name": &Config{
			name:"database_name",
			defaultValue:"planets",
			usage:"Database's name of this API's database.",
		},
		"db_password": &Config{
			name:"db_password",
			defaultValue:"e5e66467-8b00-421a-af8c-00a08e038f04",
			usage:"Password of this API's database.",
		},
		"collection_name": &Config{
			name:"collection_name",
			defaultValue:"planets",
			usage:"Collections's name of this API's database.",
		},
	}

	parseArgs(configs)
	flag.Parse()


	client := model.MongoDBConnect(
		*configs["db_host"].valuePtr,
		*configs["db_port"].valuePtr,
		*configs["db_user"].valuePtr,
		*configs["db_password"].valuePtr,
		*configs["database_name"].valuePtr,
		*configs["collection_name"].valuePtr,
	)
	defer model.MongoDBDisconnect(client)

	api.HandleRequests(*configs["host"].valuePtr, *configs["port"].valuePtr)
}
