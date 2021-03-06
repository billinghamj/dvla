package dvla

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	pkgerr "github.com/pkg/errors"
)

const confirmVehicleURI = "https://vehicleenquiry.service.gov.uk/ConfirmVehicle"
const viewVehicleURI = "https://vehicleenquiry.service.gov.uk/ViewVehicle"
const viewVehicleFormAction = "/ViewVehicle"

type viewVehicleParams struct {
	viewstate, vrm, make, colour string
}

// VehicleDetails structure to store vehicle details
type VehicleDetails struct {
	Registration      string `json:"registration"`
	Make              string `json:"make"`
	Taxed             string `json:"taxed"`
	DateRegistered    string `json:"dateRegistered"`
	YearOfManufacture string `json:"yearOfManufacture"`
	CylinderCapacity  string `json:"cylinderCapacity"`
	CO2Emissions      string `json:"co2Emissions"`
	FuelType          string `json:"fuelType"`
	ExportMarker      string `json:"exportMarker"`
	TaxStatus         string `json:"taxStatus"`
	TaxDue            string `json:"taxDue"`
	Colour            string `json:"colour"`
	Wheelplan         string `json:"wheelplan"`
	Weight            string `json:"weight"`
}

// make the first request to the DVLA to get an HTML form for the next step
func initialLookup(reg string) (*goquery.Document, error) {
	// make the request to the DVLA service
	response, err := http.PostForm(
		confirmVehicleURI,
		url.Values{
			"Vrm":      {reg},
			"Continue": {""},
		},
	)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Create a goquery document from the HTTP response
	return goquery.NewDocumentFromReader(response.Body)
}

// parse the ConfirmVehicle form for the parameters for the next stage
func getViewVehicleParamsFromPage(document *goquery.Document) (*viewVehicleParams, error) {
	var params viewVehicleParams

	// find the form element in the returned HTML
	forms := document.Find("form")

	// iterate through all the forms, in case there are more than one
	for f := range forms.Nodes {
		form := forms.Eq(f)
		// look for the form with the target action
		action, exists := form.Attr("action")
		if !exists || action != viewVehicleFormAction {
			continue
		}

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
					params.vrm = input.AttrOr("value", "")
				case "Make":
					params.make = input.AttrOr("value", "")
				case "Colour":
					params.colour = input.AttrOr("value", "")
				default:
				}
			}
		}

		break
	}

	// if any of the parameter parts are missing then return an error
	if params.viewstate == "" || params.vrm == "" || params.make == "" || params.colour == "" {
		return nil, errors.New("Failed to retrieve params for next action")
	}
	// otherwise return the complete parameters
	return &params, nil
}

// extract the formatted registration from the page
func getFormattedRegFromPage(element *goquery.Selection) string {
	h1 := element.Find("h1")
	if len(h1.Nodes) == 1 {
		return strings.TrimSpace(h1.Text())
	}

	return ""
}

func getTaxedStatusFromPage(element *goquery.Selection) (string, string) {
	var taxStatus, taxDue string

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

func getVehicleDetailsFromPage(element *goquery.Selection) (*VehicleDetails, error) {
	var details VehicleDetails

	div := element.Find("div.related-links")
	if len(div.Nodes) < 1 {
		return nil, errors.New("Unable to find vehicle details - related links missing")
	}
	lis := div.Find("li")
	for l := range lis.Nodes {
		li := lis.Eq(l)
		liText := li.Text()
		strong := li.Find("strong")
		if len(strong.Nodes) != 1 {
			return nil, errors.New("Unable to find vehicle details - item value missing")
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

	return &details, nil
}

func secondaryLookup(params *viewVehicleParams) (*goquery.Document, error) {
	response, err := http.PostForm(
		viewVehicleURI,
		url.Values{
			"viewstate": {params.viewstate},
			"Vrm":       {params.vrm},
			"Make":      {params.make},
			"Colour":    {params.colour},
			"Correct":   {"True"},
			"Continue":  {""},
		},
	)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return goquery.NewDocumentFromReader(response.Body)
}

func getVehicleDataFromPage(document *goquery.Document) (*VehicleDetails, error) {
	main := document.Find("main")

	reg := getFormattedRegFromPage(main)
	taxStatus, taxDue := getTaxedStatusFromPage(main)

	vehicleDetails, err := getVehicleDetailsFromPage(main)
	if err != nil {
		return nil, err
	}

	vehicleDetails.Registration = reg
	vehicleDetails.Taxed = taxStatus
	vehicleDetails.TaxDue = taxDue

	return vehicleDetails, nil
}

// Check lookup registration on the DVLA website
func Check(reg string) (*VehicleDetails, error) {
	// make the first request to DVLA to confirm the registration
	firstDoc, err := initialLookup(reg)
	if err != nil {
		return nil, pkgerr.Wrap(err, "Error performing initial lookup")
	}

	// get the parameters for the next request from the returned page
	params, err := getViewVehicleParamsFromPage(firstDoc)
	if err != nil {
		return nil, pkgerr.Wrap(err, "Error getting parameters for secondary lookup")
	}

	// use the form parameters to perform the secondary lookup
	secondDoc, err := secondaryLookup(params)
	if err != nil {
		return nil, pkgerr.Wrap(err, "Error performing secondary lookup")
	}

	// use the parameters to make the final request
	details, err := getVehicleDataFromPage(secondDoc)
	if err != nil {
		return nil, pkgerr.Wrap(err, "Error extracting vehicle details")
	}

	return details, nil
}
