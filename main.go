package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
	"github.com/scylladb/gocqlx/table"
)

/*Planet ... Saves planet data*/
type Planet struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Climate string `json:"climate"`
	Terrain string `json:"terrain"`
}

//TODO: Remove these global vars.
//var planets []Planet

var planetMetadata = table.Metadata{
	Name:    "planet",
	Columns: []string{"ID", "Name", "Climate", "Terrain"},
	SortKey: []string{"id"},
}

var planetTable = table.New(planetMetadata)

var session gocqlx.Session

/* Database functions */

func dbConnect() gocqlx.Session {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "planet"
	cluster.Consistency = gocql.Quorum
	_session, error := gocqlx.WrapSession(
		cluster.CreateSession(),
	)

	if error != nil {
		log.Fatal(error)
		log.Fatal("Error while trying to connect to database")
	}
	return _session
}

/*InsertPlanet adds a new planet to the database. */
func InsertPlanet(newPlanet Planet) bool {
	var created bool = true

	stmt, stmtError := planetTable.Insert()
	query := session.Query(stmt, stmtError).BindStruct(newPlanet)
	if err := query.ExecRelease(); err != nil {
		created = false
		log.Fatal(err)
		log.Fatal("Error while saving planet:", newPlanet.Name)
	}
	return created
}

/*SelectPlanetByParam gets planets from database by paramName id or name */
func SelectPlanetByParam(paramName string, paramValue ...interface{}) []Planet {
	if len(paramValue) != 1 {
		panic("Please, inform just one parameter for SelectPlanetByParam")
	}
	var planets []Planet
	value := paramValue[0]

	stmt, stmtError := planetTable.Select()
	queryMap := qb.M{paramName: value}
	query := session.Query(stmt, stmtError).BindMap(queryMap)
	execError := query.SelectRelease(&planets)
	if execError != nil {
		log.Fatal(execError)
		log.Fatal("Error while retrieving planet by", paramName, "=", paramValue)
	}

	return planets
}

/*DeletePlanetByParam delets a planet from database, informing a name or id*/
func DeletePlanetByParam(paramName string, paramValue ...interface{}) bool {
	var deleted bool = true

	stmt, stmtError := planetTable.Delete()
	queryMap := qb.M{paramName: paramValue}
	query := session.Query(stmt, stmtError).BindStruct(queryMap)
	if err := query.ExecRelease(); err != nil {
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

	return planets
}

func SearchByParam(paramName string, value ...interface{}) []Planet {
	// Acess method from db given name and returns Planet. Case insensitive
	var data []Planet
	data = SelectPlanetByParam(paramName, value)

	return data
}

func RemovePlanetByParam(paramName string, value ...interface{}) bool {
	// Removes a planet by id or name. If planet id not found, raises exception
	//var planet Planet
	removed := DeletePlanetByParam(paramName, value)

	// TODO: Remove planet in DB
	//returns it?

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

	session = dbConnect()
	defer session.Close()

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
