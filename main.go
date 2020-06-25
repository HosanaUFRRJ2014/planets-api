package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*Planet ... Saves planet data*/
type Planet struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name    string `bson:"name" json:"name"`
	Climate string `bson:"climate" json:"climate"`
	Terrain string `bson:"terrain" json:"terrain"`
}

// TODO: Remove Global Vars
var collection *mongo.Collection

/* Database functions */

func makeURI(host, user, password, databaseName string) string {
	uri := "mongodb+srv://" + user + ":" + password + "@" + host + "/" + databaseName + "?retryWrites=true&w=majority"
	return uri
}

func mongoDBConnect(host, user, password, databaseName, collectionName string) (*mongo.Client, *mongo.Collection) {
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := makeURI(host, user, password, databaseName)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))

	if err != nil {
		log.Println("Could not connect to mongo db") 
		log.Println(err)
	}

	collection := client.Database(databaseName).Collection(collectionName)
	log.Print(
		"Connected to Database: ", databaseName,
		", collection: ", collectionName,
		" at ", host,
	)

	return client, collection

}

func mongoDBDisconnect(client *mongo.Client) {
	error := client.Disconnect(context.TODO())

	if error != nil {
		log.Fatal(error)
	}
	log.Print("Disconnecting from DB.")
}

/*InsertPlanet adds a new planet to the database. Ignores planets with the same name*/
func InsertPlanet(newPlanet Planet) bool {
	var created bool
	filter := bson.D{{"name", newPlanet.Name}}
	
	// If the a planet with the same name already exists, ignore insertion
	var replaceOptions *options.ReplaceOptions = options.Replace()
	replaceOptions.SetUpsert(true)

	result , err := collection.ReplaceOne(
		context.TODO(),
		filter,
		newPlanet,
		replaceOptions,
	)
	//created := result.get("InsertedID")
	if err != nil {
		created = false
		log.Fatal(err)
		log.Fatal("Error while saving planet: ", newPlanet.Name)
	}
	log.Println("UpsertedID ID: ", result.UpsertedID)
	created = result.UpsertedID != nil
	return created
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

		if ! tempPlanet.isEmpty() {
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
	value := paramValue[0]
	if paramName == "id" {
		paramName = "_id"
		strValue := fmt.Sprintf("%v", value)
		value, _ = primitive.ObjectIDFromHex(strValue)
	}
	filter := bson.D{{paramName, value}}
	err := collection.FindOne(context.TODO(), filter).Decode(&planet)
	if err != nil {
		log.Println(
			"No planet found when", paramName, "=", value, "\n",
			err,
		)
	}
	return planet
}

/*DeletePlanetByParam delets a planet from database, informing a name or id*/
func DeletePlanetByParam(paramName string, paramValue ...interface{}) bool {
	if len(paramValue) != 1 {
		panic("Please, inform just one parameter for DeletePlanetByParam")
	}
	var deleted bool
	value := paramValue[0]
	filter := bson.D{{paramName, value}}

	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {

		log.Fatal(err)
		log.Fatal("Error while deleting planet with", paramName, "=", paramValue)
	}
	deleted = result.DeletedCount > 0
	return deleted
}

// **
func (planet Planet) movieAppearenceCount() int {
	// Get data from https://swapi.dev/api/planets
	// count movies appearence and return it

	return 0

}

/* Planet methods */

func (planet Planet) isEmpty() bool {
	return planet.Name == "";
}

func (planet Planet) isEqual(otherPlanet Planet) bool {
	//hasEqualID := planet.ID == otherPlanet.ID
	hasEqualName := strings.ToLower(planet.Name) == strings.ToLower(planet.Name)

	return hasEqualName 
}

/* Model Functions */

func addNewPlanet(newPlanet Planet) bool {
	//TODO: Check if planet is already in database
	created := InsertPlanet(newPlanet)

	return created
}

func getAllPlanets() []Planet {
	// Retrieve all planets from database and returns it
	var planets []Planet
	planets = GetPlanetsFromDB()
	// TODO: Para cada planetadeve retornar também a quantidade de aparições
	// em filmes


	return planets
}

//func SearchByParam(paramName string, value ...interface{}) []Planet {
func SearchByParam(paramName string, value ...interface{}) Planet {
	// Acess method from db given name and returns Planet. Case insensitive
	var planet Planet
	planet = SelectPlanetByParam(paramName, value[0])

	return planet
}

func RemovePlanetByParam(paramName string, value ...interface{}) bool {
	// Removes a planet by id or name. If planet id not found, raises exception
	//var planet Planet
	removed := DeletePlanetByParam(paramName, value[0])
	//returns removed planet?

	return removed
}

/* API Utils*/

func capitalizeName(name string) string {
	var capitalizedName string

	if len(name) >= 1 {
		capitalizedName = strings.ToUpper(name[0:1]) + strings.ToLower(name[1:])
	} else {
		capitalizedName = strings.ToUpper(name)
	}

	return capitalizedName
}

func getByAttribute(attributeName string, request *http.Request) string {
	variables := mux.Vars(request)
	return variables[attributeName]
}

func parseIDToInt64(id string) int64 {
	idAsInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Fatal("ID ", id, " is not of type integer")
	}

	return idAsInt
}

func formatResponse(writer *http.ResponseWriter, data ...interface{}) {
	encoder := json.NewEncoder(*writer)
	encoder.SetIndent("", "\t")

	if data[0] == nil {
		//removes nil from response
		data = data[:len(data)-1] 
	}
	encoder.Encode(data)
}

func formatPlanetResponse(writer *http.ResponseWriter, planet Planet) {
	if planet.isEmpty() {
		formatResponse(writer, nil)
	} else {
		formatResponse(writer, planet)
	}
}

/*API functions*/

func APIHome(writer http.ResponseWriter, request *http.Request) {
	homeTemplate, _ := template.ParseFiles("home.html")
	homeTemplate.ExecuteTemplate(writer, "home.html", nil)
}

func CreateNewPlanet(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Fatal("Could not read body for request")
		panic(err)
	}

	var newPlanet Planet
	json.Unmarshal(body, &newPlanet)
	newPlanet.Name = capitalizeName(newPlanet.Name)
	created := addNewPlanet(newPlanet)

	// TODO: Return new id and new created object?
	response := map[string]bool{"created": created}

	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if created {
		writer.WriteHeader(http.StatusCreated)
	} else {
		writer.WriteHeader(http.StatusBadRequest)
	}
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	encoder.Encode(response)
}

func ListPlanets(writer http.ResponseWriter, request *http.Request) {
	planets := getAllPlanets()

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	if len(planets) == 0 {
		empty := []int{}
		encoder.Encode(empty)
	} else {
		encoder.Encode(planets)
	}

}

func GetByID(writer http.ResponseWriter, request *http.Request) {
	param := "id"
	id := getByAttribute(param, request)
	//idAsInt := parseIDToInt64(id)
	planet := SearchByParam(param, id)

	formatPlanetResponse(&writer, planet)

}

func GetByName(writer http.ResponseWriter, request *http.Request) {
	param := "name"
	name := getByAttribute(param, request)
	capitalizedName := capitalizeName(name)
	planet := SearchByParam(param, capitalizedName)

	formatPlanetResponse(&writer, planet)
}

func DeletePlanetByName(writer http.ResponseWriter, request *http.Request) {
	paramName := "name"
	name := getByAttribute(paramName, request)
	capitalizedName := capitalizeName(name)

	deleted := RemovePlanetByParam(paramName, capitalizedName)
	response := map[string]bool{"deleted": deleted}
	formatResponse(&writer, response)
}

func DeletePlanetByID(writer http.ResponseWriter, request *http.Request) {
	paramName := "id"
	id := getByAttribute(paramName, request)
	//idAsInt := parseIDToInt64(id)

	deleted := RemovePlanetByParam(paramName, id)
	response := map[string]bool{"deleted": deleted}
	formatResponse(&writer, response)
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", APIHome)
	router.HandleFunc("/planets", ListPlanets).Methods("GET")
	router.HandleFunc("/planet", CreateNewPlanet).Methods("POST")
	router.HandleFunc("/planet/id/{id}", GetByID).Methods("GET")
	router.HandleFunc("/planet/name/{name}", GetByName).Methods("GET")
	router.HandleFunc("/planet/id/{id}", DeletePlanetByID).Methods("DELETE")
	router.HandleFunc("/planet/name/{name}", DeletePlanetByName).Methods("DELETE")

	log.Println("Listening at port 5555...")
	log.Fatal(http.ListenAndServe(":5555", router))
}

func main() {

	var (
		host string = "planetsapi-6dhqu.gcp.mongodb.net"
		user string = "planetsapi"
		password string = "940bf335-e572-496e-af0e-10ec3b9e1cd3"
		databaseName string = "planetsapi"
		collectionName string = "planets"
	)

	var client *mongo.Client
	client, collection = mongoDBConnect(
		host,
		user,
		password,
		databaseName,
		collectionName,
	)
	defer mongoDBDisconnect(client)

	handleRequests()
}
