package planet

import (
	"github.com/HosanaUFRRJ2014/planets-api/model"
)

/* Functions */

func AddNewPlanet(newPlanet model.Planet) bool {
	
	var created bool = false

	if !newPlanet.IsEmpty() {
		created = model.InsertPlanet(newPlanet)
	}

	return created
}

func GetAllPlanets() []model.Planet {
	// Retrieve all planets from database and returns it
	var planets []model.Planet
	planets = model.GetPlanetsFromDB()

	return planets
}

// Acess method from db given name and returns Planet. Case insensitive
func SearchByParam(paramName string, value ...interface{}) model.Planet {
	var planet model.Planet
	planet = model.SelectPlanetByParam(paramName, value[0])

	return planet
}

// Removes a planet by id or name. If planet param not found, returns false
func RemovePlanetByParam(paramName string, value ...interface{}) bool {

	removed := model.DeletePlanetByParam(paramName, value[0])
	//TODO: returns removed planet?

	return removed
}