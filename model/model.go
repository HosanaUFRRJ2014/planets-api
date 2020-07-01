package model

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


/*Planet ... Saves planet data in our API*/
type Planet struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name    string `bson:"name" json:"name"`
	Climate string `bson:"climate" json:"climate"`
	Terrain string `bson:"terrain" json:"terrain"`
	AppearencesCount int `bson:"appearencesCount" json:"appearencesCount"`
	PlanetSwapiURL string `bson:"planetSwapiURL" json:"-"`
}


/* Planet methods */

func (planet Planet) IsEmpty() bool {
	return planet.Name == "";
}

var collection *mongo.Collection


/* Database functions */

func makeURI(host, port, databaseName string) string {
	uri := "mongodb" + "://" + host + ":" + port + "/" + databaseName + "?retryWrites=true&w=majority"
	return uri
}

func MongoDBConnect(host, port, user, password, databaseName, collectionName string) (*mongo.Client, /**mongo.Collection*/) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	credential := options.Credential{
		Username: user,
		Password: password,
		AuthMechanism: "",
	}
	uri := makeURI(host, port, databaseName)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetAuth(credential))

	if err != nil {
		log.Println("Could not connect to mongo db") 
		log.Println(err)
	}

	database := client.Database(databaseName)
	collection = database.Collection(collectionName)

	// Ensure that two data keys cannot share the same name.
	nameIndex := mongo.IndexModel{
		Keys: bson.D{{"name", 1}},
		Options: options.Index().
			SetUnique(true).
			SetPartialFilterExpression(bson.D{
				{"name", bson.D{
					{"$exists", true},
				}},
			}),
	}
	if _, err = collection.Indexes().CreateOne(context.TODO(), nameIndex); err != nil {
		log.Println("Erro while saving duplicated planet. Ignoring")
	}

	log.Print(
		"Connected to Database: ", databaseName,
		", collection: ", collectionName,
		" at ", host,
	)

	return client//, collection

}

func MongoDBDisconnect(client *mongo.Client) {
	error := client.Disconnect(context.TODO())

	if error != nil {
		log.Fatal(error)
	}
	log.Print("Disconnecting from DB.")
}

/*InsertPlanet adds a new planet to the database. Ignores planets with the same name*/
func InsertPlanet(newPlanet Planet) (bool, string, string) {
	var created bool = false
	var errorMessage string = ""
	var planetUUID string
	result, err := collection.InsertOne(context.TODO(), newPlanet)
	
	if err != nil {
		created = false
		log.Println("Error while saving planet: " + newPlanet.Name)
	} else {
		created = result.InsertedID != nil
		planetUUID = result.InsertedID.(primitive.ObjectID).Hex()
	}
	if !created {
		errorMessage = "Planet " + newPlanet.Name + " already exists"
	}

	return created, planetUUID, errorMessage
}

func GetPlanetsFromDB() []Planet {
	var planets [] Planet
	emptyFilter := bson.D{{}}

	cursor, err := collection.Find(context.TODO(), emptyFilter)
	if err != nil {
		log.Fatal(err)
		log.Fatal("Error while retrieving all planets")
	}

	// Parsing list of planets
	for cursor.Next(context.TODO()) {
		var tempPlanet Planet
		err := cursor.Decode(&tempPlanet)
		if err != nil {
			log.Fatal(err)
			log.Fatal("Could not parse list of planets")
		}

		if ! tempPlanet.IsEmpty() {
			planets = append(planets, tempPlanet)
		}
	}

	return planets
}

/*SelectPlanetByParam gets planets from database by paramName id or name */
func SelectPlanetByParam(paramName string, paramValue ...interface{}) Planet {
	if len(paramValue) != 1 {
		panic("Please, inform just one parameter for SelectPlanetByParam")
	}
	var planet Planet
	var dbParamName string
	value := paramValue[0]
	if paramName == "id" {
		dbParamName = "_id"
		strValue := fmt.Sprintf("%v", value)
		value, _ = primitive.ObjectIDFromHex(strValue)
	} else {
		dbParamName = paramName
	}
	filter := bson.D{{dbParamName, value}}
	err := collection.FindOne(context.TODO(), filter).Decode(&planet)
	if err != nil {
		log.Println(
			"Error while retrieving planet with " + paramName + " = " + paramValue[0].(string), "\n",
			err,
		)
	}

	return planet
}

/*DeletePlanetByParam delets a planet from database, informing a name or id*/
func DeletePlanetByParam(paramName string, paramValue ...interface{}) (bool, string) {
	if len(paramValue) != 1 {
		panic("Please, inform just one parameter for DeletePlanetByParam")
	}
	var deleted bool = false
	var errorMessage string = ""
	var dbParamName string
	value := paramValue[0]
	if paramName == "id" {
		dbParamName = "_id"
		strValue := fmt.Sprintf("%v", value)
		value, _ = primitive.ObjectIDFromHex(strValue)
	} else {
		dbParamName = paramName
	}
	filter := bson.D{{dbParamName, value}}

	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		errorMessage = "Error while deleting planet with " + paramName + " = " + paramValue[0].(string)
		log.Fatal(err)
		log.Fatal(errorMessage)
	} else {
		deleted = result.DeletedCount > 0
		if !deleted {
			errorMessage = "Planet with " + paramName + " = " + paramValue[0].(string) + " not found"
		}
	}
	return deleted, errorMessage
}