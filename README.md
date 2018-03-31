# dvla

dvla is a library which provides vehicle registration lookup via the DVLA GOV.UK website. 

## Table of Contents

* [Installation](#installation)
* [Changelog](#changelog)
* [API](#api)
* [Example](#example)

## Installation

Please note that because of the goquery dependency, dvla requires Go1.1+.

    $ go get github.com/YodaTheCoder/dvla

## Changelog

*    **2018-03-27 (v0.1.0)** : Initial commit, very much WIP.

## API

dvla exposes one struct `VehicleDetails` which is returned by the one interface `Check`.

## Example

```Go
package main

import (
	"log"

	"github.com/YodaTheCoder/dvla"
)

func main() {
	log.Println("regLookup")
	vehicleDetails := dvla.Check("MT07TYW")

	log.Printf("Make: '%s'", vehicleDetails.Make)
	log.Printf("Colour: '%s'", vehicleDetails.Colour)
	log.Printf("Taxed: '%s'", vehicleDetails.Taxed)
	log.Printf("Tax Status: '%s'", vehicleDetails.TaxStatus)
	log.Printf("Tax Due: '%s'", vehicleDetails.TaxDue)
	log.Printf("Date Registered: '%s'", vehicleDetails.DateRegistered)
	log.Printf("Year of Manufacture: '%s'", vehicleDetails.YearOfManufacture)
	log.Printf("Cylinder Capacity: '%s'", vehicleDetails.CylinderCapacity)
	log.Printf("CO₂ Emissions: '%s'", vehicleDetails.CO2Emissions)
	log.Printf("Fuel Type: '%s'", vehicleDetails.FuelType)
	log.Printf("Export Marker: '%s'", vehicleDetails.ExportMarker)
	log.Printf("Wheelplan: '%s'", vehicleDetails.Wheelplan)
	log.Printf("Weight: '%s'", vehicleDetails.Weight)
}
```

