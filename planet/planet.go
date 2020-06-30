package planet

import (
	"strings"
	"github.com/HosanaUFRRJ2014/planets-api/model"
)


func capitalizeName(name string) string {
	var capitalizedName string

	if len(name) >= 1 {
		capitalizedName = strings.ToUpper(name[0:1]) + strings.ToLower(name[1:])
	} else {
		capitalizedName = strings.ToUpper(name)
	}

	return capitalizedName
}

/*Applies PrepareString if request is search or delete by name*/
func prepareParam(paramName string, value ...interface{}) (string, string) {
	var searcheableValue string
	var errorMessage string = ""

	if paramName == "name" {
		searcheableValue, errorMessage = PrepareString(value[0].(string))
	} else {
		searcheableValue = value[0].(string)
	}

	return searcheableValue, errorMessage
}

/*Applies trim by space and capitalization*/
func PrepareString(name string) (string, string) {
	trimmedName := strings.Trim(name, " ")
	capitalizedName := capitalizeName(trimmedName)

	errorMessage := ""
	if len(capitalizedName) == 0 {
		errorMessage = "Could not do action for empty param"
	}

	return capitalizedName, errorMessage
}


/* Functions */

/*Creates new planet*/
func AddNewPlanet(newPlanet model.Planet) (bool, string) {
	
	var created bool = false
	var errorMessage string = ""

	if !newPlanet.IsEmpty() {
		newPlanet.Name, errorMessage = PrepareString(newPlanet.Name)

		if errorMessage == "" {
			created, errorMessage = model.InsertPlanet(newPlanet)
		}
	}

	return created, errorMessage
}

/* Retrieve all planets from database and returns it*/
func GetAllPlanets() []model.Planet {
	var planets []model.Planet
	planets = model.GetPlanetsFromDB()

	return planets
}

// Searches planet by id or name and returns Planet. Case insensitive
func SearchByParam(paramName string, value ...interface{}) (model.Planet, string) {
	var planet model.Planet

	searcheableValue, errorMessage := prepareParam(paramName, value[0])

	if errorMessage == "" {
		planet = model.SelectPlanetByParam(paramName, searcheableValue)
	}

	return planet, errorMessage
}

// Removes a planet by id or name. If planet param not found, returns false
func RemovePlanetByParam(paramName string, value ...interface{}) (bool, string) {
	var removed bool = false
	var errorMessage string = ""
	removableValue, errorMessage := prepareParam(paramName, value[0]) 

	if errorMessage == "" {
		removed, errorMessage = model.DeletePlanetByParam(paramName, removableValue)
	}
	//TODO: returns removed planet?

	return removed, errorMessage
}