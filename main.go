package main

import (
	"context"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*Planet ... Saves planet data*/
type Planet struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Climate string `json:"climate"`
	Terrain string `json:"terrain"`
}

// TODO: Remove Global Vars
var collection *mongo.Collection

/* Database functions */

func makeURI(host, port string) string {
	driverName := "mongodb"
	uri := driverName + "://" + host + ":" + port

	return uri
}

func dbConnect(databaseName, collectionName string) (*mongo.Client, *mongo.Collection) {
	var (
		host string = "localhost"
		port string = "27017"
	)

	uri := makeURI(host, port)
	var clientOptions options.ClientOptions
	client, error := mongo.NewClient(clientOptions.ApplyURI(uri))
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	error = client.Connect(ctx)

	if error != nil {
		log.Fatal(error)
	}

	collection := client.Database(databaseName).Collection(collectionName)
	log.Print(
		"Connected to Database: ", databaseName,
		", collection: ", collectionName,
		" at ", uri,
	)

	return client, collection

}

func dbDisconnect(client *mongo.Client) {
	error := client.Disconnect(context.TODO())

	if error != nil {
		log.Fatal(error)
	}
	log.Print("Disconnecting from DB.")
}

/*InsertPlanet adds a new planet to the database. */
func InsertPlanet(newPlanet Planet) bool {
	created := true
	_, err := collection.InsertOne(context.TODO(), newPlanet)
	//created := result.get("InsertedID")
	if err != nil {
		created = false
		log.Fatal(err)
		log.Fatal("Error while saving planet:", newPlanet.Name)
	}
	return created
}

/*SelectPlanetByParam gets planets from database by paramName id or name */
func SelectPlanetByParam(paramName string, paramValue ...interface{}) Planet {
	if len(paramValue) != 1 {
		panic("Please, inform just one parameter for SelectPlanetByParam")
	}
	var planet Planet
	value := paramValue[0]
	filter := bson.D{{paramName, value}}

	err := collection.FindOne(context.TODO(), filter).Decode(&planet)
	if err != nil {
		log.Fatal(err)
		log.Fatal("Error while retrieving planet by", paramName, "=", paramValue)
	}

	return planet
}

/*DeletePlanetByParam delets a planet from database, informing a name or id*/
func DeletePlanetByParam(paramName string, paramValue ...interface{}) bool {
	if len(paramValue) != 1 {
		panic("Please, inform just one parameter for SelectPlanetByParam")
	}
	deleted := true
	value := paramValue[0]
	filter := bson.D{{paramName, value}}

	_, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		deleted = false
		log.Fatal(err)
		log.Fatal("Error while deleting planet with", paramName, "=", paramValue)
	}
	return deleted
}

// **
func (planet Planet) movieAppearenceCount() int {
	// Get data from https://swapi.dev/api/planets
	// count movies appearence and return it

	return 0

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
	//planets := GetPlanetsFromDB()

	// Para cada planetadeve retornar também a quantidade de aparições em filmes
	panic("getAllPlanets not implemented yet")
	return planets
}

//func SearchByParam(paramName string, value ...interface{}) []Planet {
func SearchByParam(paramName string, value ...interface{}) Planet {
	// Acess method from db given name and returns Planet. Case insensitive
	//var data []Planet
	var data Planet
	data = SelectPlanetByParam(paramName, value)

	return data
}

func RemovePlanetByParam(paramName string, value ...interface{}) bool {
	// Removes a planet by id or name. If planet id not found, raises exception
	//var planet Planet
	removed := DeletePlanetByParam(paramName, value)
	//returns removed planet?

	return removed
}

/* API Utils*/

func capitalizeName(name string) string {
	var capitalizedName string

	capitalizedName = strings.ToTitle(
		strings.ToLower(name),
	)

	return capitalizedName
}

func getByAttribute(attributeName string, request *http.Request) string {
	variables := mux.Vars(request)
	return variables[attributeName]
}

func parseIDToInt64(id string) int64 {
	idAsInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Fatal("ID", id, "is not of type integer")
	}

	return idAsInt
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
	encoder.Encode(planets)
}

func GetByID(writer http.ResponseWriter, request *http.Request) {
	param := "id"
	id := getByAttribute(param, request)
	idAsInt := parseIDToInt64(id)
	data := SearchByParam(param, idAsInt)

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	encoder.Encode(data)

}

func GetByName(writer http.ResponseWriter, request *http.Request) {
	param := "name"
	name := getByAttribute(param, request)
	capitalizedName := capitalizeName(name)
	data := SearchByParam(param, capitalizedName)

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	encoder.Encode(data)
}

func DeletePlanetByName(writer http.ResponseWriter, request *http.Request) {
	paramName := "name"
	name := getByAttribute(paramName, request)
	capitalizedName := capitalizeName(name)
	deleted := RemovePlanetByParam(paramName, capitalizedName)

	response := map[string]bool{"deleted": deleted}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	encoder.Encode(response)
}

func DeletePlanetByID(writer http.ResponseWriter, request *http.Request) {
	paramName := "id"
	id := getByAttribute(paramName, request)
	idAsInt := parseIDToInt64(id)

	deleted := RemovePlanetByParam(paramName, idAsInt)

	response := map[string]bool{"deleted": deleted}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	encoder.Encode(response)
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", APIHome)
	router.HandleFunc("/planets", ListPlanets)
	router.HandleFunc("/planet", CreateNewPlanet).Methods("POST")
	router.HandleFunc("/planet/id/{id}", GetByID)
	router.HandleFunc("/planet/name/{name}", GetByName)
	router.HandleFunc("/planet/{id}", DeletePlanetByID).Methods("DELETE")

	log.Println("Listening at port 5555...")
	log.Fatal(http.ListenAndServe(":5555", router))
}

func main() {

	var (
		databaseName   string = "apidb"
		collectionName string = "planets"
	)

	var client *mongo.Client
	client, collection = dbConnect(databaseName, collectionName)
	defer dbDisconnect(client)

	// planets = []Planet{
	// 	Planet{ID: 0, Name: "Tatooine", Climate: "hot", Terrain: "sand"},
	// 	Planet{ID: 1, Name: "Earth", Climate: "sunnie", Terrain: "rocks"},
	// }

	//	planets := ListPlanets()

	//	fmt.Println(planets)

	//p := Planet{3, "Mars", "hot", "rocks"}
	//addNewPlanet(p)

	//fmt.Println(planets)

	handleRequests()
}
