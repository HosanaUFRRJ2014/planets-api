package api


import (
	"encoding/json"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"fmt"
	"strings"

	"github.com/gorilla/mux"

	"github.com/HosanaUFRRJ2014/planets-api/model"
	"github.com/HosanaUFRRJ2014/planets-api/planet"
)

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

const PLANETS_SWAPI_URL string = "https://swapi.dev/api/planets/"

var ERROR_MESSAGES = map[string]string{
	"WRONG_PARAMS": "Query param is invalid. Valid options: ?id= , ?name= ",
	//""
}


/* API Utils*/
func getByAttribute(request *http.Request) (string, string, string) {
	variables := mux.Vars(request)
	var errorMessage string = ""

	paramName := "name"
	paramValue, ok := variables["name"]

	if !ok {
		paramValue, ok = variables["id"]
		paramName = "id"
	}

	if !ok {
		errorMessage = "Query param is invalid. Valid options: ?id= , ?name= "
	}

	return paramName, paramValue, errorMessage
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

func formatPlanetResponse(writer *http.ResponseWriter, planet model.Planet) {
	if planet.IsEmpty() {
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

func getAppearencesCountFromSWAPI(planetName string) (int, string) {
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
	homeTemplate, error := template.ParseFiles("home.html", "static/css/apistyle.css")
	
	if error != nil {
		panic(error)
	}
	
	homeTemplate.ExecuteTemplate(writer, "home.html", nil)
}

func CreateNewPlanet(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Fatal("Could not read body for request")
	}

	var newPlanet model.Planet
	var created bool = false
	var errorMessage string = ""
	var response map[string]interface{}
	json.Unmarshal(body, &newPlanet)

	newPlanet.Name, errorMessage = planet.PrepareString(newPlanet.Name)

	if errorMessage == "" {
		// Get appearences count
		appearencesCount, planetSwapiURL := getAppearencesCountFromSWAPI(newPlanet.Name)
	
		//Updates new planet with swapi information
		newPlanet.AppearencesCount = appearencesCount
		newPlanet.PlanetSwapiURL = planetSwapiURL
	
		// Saving Planet
		created, errorMessage = planet.AddNewPlanet(newPlanet)
	}

	// TODO: Return new id and new created object?

	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if created {
		response = map[string]interface{}{"created": created}
		writer.WriteHeader(http.StatusCreated)
	} else {
		response = map[string]interface{}{"created": created, "error": errorMessage}
		writer.WriteHeader(http.StatusBadRequest)
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	encoder.Encode([1]map[string]interface{}{response})
}

func ListPlanets(writer http.ResponseWriter, request *http.Request) {
	planets := planet.GetAllPlanets()

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	if len(planets) == 0 {
		empty := []int{}
		encoder.Encode(empty)
	} else {
		encoder.Encode(planets)
	}

}

func GetByParam(writer http.ResponseWriter, request *http.Request) {
	var errorMessage string = ""
	var retrievedPlanet model.Planet;
	paramName, paramValue, errorMessage := getByAttribute(request)

	if errorMessage == "" {
		retrievedPlanet, errorMessage = planet.SearchByParam(paramName, paramValue)

		if errorMessage == "" {
			formatPlanetResponse(&writer, retrievedPlanet)
		}
	} 
	
	if len(errorMessage) > 0 {
		response := map[string]string{"error": errorMessage}
		formatResponse(&writer, response)
	}
}

func DeletePlanetByParam(writer http.ResponseWriter, request *http.Request) {
	var errorMessage string = ""
	var deleted bool = false
	var response map[string]interface{}
 
	paramName, paramValue, errorMessage := getByAttribute(request)

	if errorMessage == "" {
		deleted, errorMessage = planet.RemovePlanetByParam(paramName, paramValue)
	} 

	if deleted {
		response = map[string]interface{}{"deleted": deleted}
	} else {
		response = map[string]interface{}{"error": errorMessage, "deleted": deleted}
	}

	formatResponse(&writer, response)
}

func HandleRequests(host, port string) {
	var dir string
	flag.StringVar(&dir, ".", "static/", "")
    flag.Parse()
	
	// Create the router
	router := mux.NewRouter().StrictSlash(true)
	
	apiRoot := "/planets/api"
	
	// Serve static Files under http://{HOST}<:{PORT}>/static/<filename>
    router.PathPrefix(apiRoot + "/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))

	// Add routers
	router.HandleFunc("/", APIHome)
	router.HandleFunc(apiRoot, APIHome)
	router.HandleFunc(apiRoot + "/planets", ListPlanets).Methods("GET")
	router.HandleFunc(apiRoot  + "/create", CreateNewPlanet).Methods("POST")
	router.Path(apiRoot + "/search").Queries("id", "{id}").HandlerFunc(GetByParam).Name("Search").Methods("GET")
	router.Path(apiRoot +  "/search").Queries("name", "{name}").HandlerFunc(GetByParam).Name("Search").Methods("GET")
	router.Path(apiRoot + "/delete").Queries("id", "{id}").HandlerFunc(DeletePlanetByParam).Name("Delete").Methods("DELETE")
	router.Path(apiRoot + "/delete").Queries("name", "{name}").HandlerFunc(DeletePlanetByParam).Name("Delete").Methods("DELETE")

	// List all paths
	err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}
		fmt.Println()
		return nil
	})

	if err != nil {
		fmt.Println(err)
	}

	http.Handle("/", router)
	
	address := host
	if len(port) > 0 {
		address = address + ":" + port
	}
	service := &http.Server{
		Handler:      router,
        Addr:         address,
        // Enforce timeouts for server
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
	}
	
	log.Println("Listening at host", host,  ", port", port, "...")
	log.Fatal(service.ListenAndServe())
}