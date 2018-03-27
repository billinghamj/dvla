package dvla

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type viewVehicleParams struct {
	viewstate, Vrm, Make, Colour, Correct, Continue string
}

// VehicleDetails structure to store vehicle details
type VehicleDetails struct {
	Registration, Make, Taxed, DateRegistered, YearOfManufacture, CylinderCapacity, CO2Emissions, FuelType, ExportMarker, TaxStatus, TaxDue, Colour, Wheelplan, Weight string
}

// make the first request to the DVLA to get a session state and confirm the registration
func initialLookup(reg string) (viewVehicleParams, error) {
	var lookupURI = "https://vehicleenquiry.service.gov.uk/ConfirmVehicle"
	var params = viewVehicleParams{Correct: "False", Continue: ""}

	// make the request to the DVLA service
	response, err := http.PostForm(
		lookupURI,
		url.Values{
			"Vrm":      {reg},
			"Continue": {""},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	// Create a goquery document from the HTTP response
	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}

	// find the form element in the returned HTML
	forms := document.Find("form")

	// iterate through all the forms, in case there are more than one
	for f := range forms.Nodes {
		form := forms.Eq(f)
		// look for the form with the target action
		action, exists := form.Attr("action")
		if exists && action == "/ViewVehicle" {
			inputs := form.Find("input")
			// iterate through all the form inputs
			for i := range inputs.Nodes {
				input := inputs.Eq(i)
				id, exists := input.Attr("id")
				// look for specific target input ids
				if exists {
					switch id {
					case "viewstate":
						params.viewstate = input.AttrOr("value", "")
					case "Vrm":
						params.Vrm = input.AttrOr("value", "")
					case "Make":
						params.Make = input.AttrOr("value", "")
					case "Colour":
						params.Colour = input.AttrOr("value", "")
					default:
					}
				}
			}
		}
	}

	// if any of the parameter parts are missing then return an error
	if params.viewstate == "" || params.Vrm == "" || params.Make == "" || params.Colour == "" {
		return params, errors.New("Failed to retrieve params for next action")
	}
	// otherwise return the complete parameters
	params.Correct = "True"
	return params, nil
}

// extract the formatted registration from the page
func getFormattedRegFromPage(element *goquery.Selection) string {
	var reg = ""
	h1 := element.Find("h1")
	if len(h1.Nodes) == 1 {
		reg = h1.Text()
	}
	return strings.TrimSpace(reg)
}

func getTaxedStatusFromPage(element *goquery.Selection) (string, string) {
	var taxStatus = ""
	var taxDue = ""

	halfDiv := element.Find("div.column-half")

	for d := range halfDiv.Nodes {
		div := halfDiv.Eq(d)
		if strings.Contains(div.Text(), "Incorrect tax status?") {
			h2 := div.Find("h2.heading-large")
			if len(h2.Nodes) == 1 {
				taxStatus = strings.TrimSpace(h2.Text())
			}
			p := div.Find("p.margin-bottom-1")
			if len(p.Nodes) == 1 {
				taxDue = strings.Replace(strings.TrimSpace(p.Text()), "Tax due:", "", 1)
			}
		}
	}
	return taxStatus, taxDue
}

func getVehicleDetailsFromPage(element *goquery.Selection) VehicleDetails {
	var details VehicleDetails

	div := element.Find("div.related-links")
	if len(div.Nodes) < 1 {
		return details
	}
	lis := div.Find("li")
	for l := range lis.Nodes {
		li := lis.Eq(l)
		liText := li.Text()
		strong := li.Find("strong")
		if len(strong.Nodes) != 1 {
			return details
		}
		switch {
		case strings.Index(liText, "Vehicle make") > -1:
			details.Make = strings.TrimSpace(strong.Text())
		case strings.Index(liText, "Date of first registration") > -1:
			details.DateRegistered = strings.TrimSpace(strong.Text())
		case strings.Index(liText, "Year of manufacture") > -1:
			details.YearOfManufacture = strings.TrimSpace(strong.Text())
		case strings.Index(liText, "Cylinder capacity (cc)") > -1:
			details.CylinderCapacity = strings.TrimSpace(strong.Text())
		case strings.Index(liText, "CO₂Emissions") > -1:
			details.CO2Emissions = strings.TrimSpace(strong.Text())
		case strings.Index(liText, "Fuel type") > -1:
			details.FuelType = strings.TrimSpace(strong.Text())
		case strings.Index(liText, "Export marker") > -1:
			details.ExportMarker = strings.TrimSpace(strong.Text())
		case strings.Index(liText, "Vehicle status") > -1:
			details.TaxStatus = strings.TrimSpace(strong.Text())
		case strings.Index(liText, "Vehicle colour") > -1:
			details.Colour = strings.TrimSpace(strong.Text())
		case strings.Index(liText, "Wheelplan") > -1:
			details.Wheelplan = strings.TrimSpace(strong.Text())
		case strings.Index(liText, "Revenue weight") > -1:
			details.Weight = strings.TrimSpace(strong.Text())
		}
	}

	return details
}

func viewVehicle(params viewVehicleParams) VehicleDetails {
	var lookupURI = "https://vehicleenquiry.service.gov.uk/ViewVehicle"

	response, err := http.PostForm(
		lookupURI,
		url.Values{
			"viewstate": {params.viewstate},
			"Vrm":       {params.Vrm},
			"Make":      {params.Make},
			"Colour":    {params.Colour},
			"Correct":   {params.Correct},
			"Continue":  {""},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	// Create a goquery document from the HTTP response
	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}

	main := document.Find("main")

	var reg = getFormattedRegFromPage(main)
	taxStatus, taxDue := getTaxedStatusFromPage(main)
	vehicleDetails := getVehicleDetailsFromPage(main)
	vehicleDetails.Registration = reg
	vehicleDetails.Taxed = taxStatus
	vehicleDetails.TaxDue = taxDue
	return vehicleDetails
}

// Check lookup registration on the DVLA website
func Check(reg string) VehicleDetails {
	var vehicle VehicleDetails

	params, err := initialLookup(reg)
	if err != nil {
		log.Fatal("Error performing initial lookup: ", err)
	}

	vehicle = viewVehicle(params)

	return vehicle
}
