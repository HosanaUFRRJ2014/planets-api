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

/*Planet ... Saves planet data in our API*/
type Planet struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name    string `bson:"name" json:"name"`
	Climate string `bson:"climate" json:"climate"`
	Terrain string `bson:"terrain" json:"terrain"`
	AppearencesCount int `bson:"appearencesCount" json:"appearencesCount"`
	PlanetSwapiURL string `bson:"planetSwapiURL" json:"-"`
}

/*Structure of a response from SWAPI planets endpoint */
type SWAPIResponse struct {
	Next string `json:"next"`
	Planets [] SWAPIPlanet `json:"results"`
}

/*Structure of planet from SWAPI that matters*/
type SWAPIPlanet struct {
	Name    string `json:"name"`
	URL string `json:"url"`
	Films [] string `json:"films"`
}

// TODO: Remove Global Vars
var collection *mongo.Collection
const PLANETS_SWAPI_URL string = "https://swapi.dev/api/planets/"

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
	result, err := collection.InsertOne(context.TODO(), newPlanet)

	if err != nil {
		created = false
		//log.Println(err)
		log.Println("Error while saving planet: ", newPlanet.Name)
	} else {
		created = result.InsertedID != nil
		log.Println("InsertedID ID: ", result.InsertedID)
	}
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
	if paramName == "id" {
		paramName = "_id"
		strValue := fmt.Sprintf("%v", value)
		value, _ = primitive.ObjectIDFromHex(strValue)
	}
	filter := bson.D{{paramName, value}}

	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {

		log.Fatal(err)
		log.Fatal("Error while deleting planet with", paramName, "=", paramValue)
	}
	deleted = result.DeletedCount > 0
	return deleted
}

/* Planet methods */

func (planet Planet) isEmpty() bool {
	return planet.Name == "";
}

/* Model Functions */

func addNewPlanet(newPlanet Planet) bool {
	
	var created bool = false

	if newPlanet.Name != "" {
		created = InsertPlanet(newPlanet)
	}

	return created
}

func getAllPlanets() []Planet {
	// Retrieve all planets from database and returns it
	var planets []Planet
	planets = GetPlanetsFromDB()

	return planets
}

// Acess method from db given name and returns Planet. Case insensitive
func SearchByParam(paramName string, value ...interface{}) Planet {
	var planet Planet
	planet = SelectPlanetByParam(paramName, value[0])

	return planet
}

// Removes a planet by id or name. If planet param not found, returns false
func RemovePlanetByParam(paramName string, value ...interface{}) bool {

	removed := DeletePlanetByParam(paramName, value[0])
	//TODO: returns removed planet?

	return removed
}

/* API Utils*/

/*Applies trim by space and capitalization*/
func prepareString(name string) string {
	trimmedName := strings.Trim(name, " ")
	capitalizedName := capitalizeName(trimmedName)

	return capitalizedName
}

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

func getSWAPIResponse(url, planetName string) SWAPIResponse {

	searchParam := "?search=" + planetName
	url = url + searchParam
	response, err := http.Get(url)

	if err != nil {
		log.Println(err.Error())
		log.Println("Error: Could not connect to SWAPI!")
	}

	responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err.Error())
	}
	
	var responseObject SWAPIResponse
	json.Unmarshal(responseData, &responseObject)

	return responseObject
}

/*API functions*/

func GetAppearencesCountFromSWAPI(planetName string) (int, string) {
	var appearencesCount int = 0
	var planetURL string = ""
	
	responseObject := getSWAPIResponse(PLANETS_SWAPI_URL, planetName)

	swapiPlanets := responseObject.Planets

	if len(swapiPlanets) > 0 {
		swapiPlanet := swapiPlanets[0]
		films := swapiPlanet.Films
		if len(films) > 0 {
			appearencesCount = len(films)
		}
		planetURL = swapiPlanet.URL

	}

	return appearencesCount, planetURL
}

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

	newPlanet.Name = prepareString(newPlanet.Name)
	// Get appearences count
	appearencesCount, planetSwapiURL := GetAppearencesCountFromSWAPI(newPlanet.Name)

	//Updates new planet with swapi information
	newPlanet.AppearencesCount = appearencesCount
	newPlanet.PlanetSwapiURL = planetSwapiURL

	// Saving Planet
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

	const (
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
