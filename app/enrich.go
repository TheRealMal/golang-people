package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type EnrichedData struct {
	Name        string
	Surname     string
	Patronymic  string
	Age         int
	Gender      string
	Nationality string
}

type AgeBody struct {
	Count int    `json:"count"`
	Name  string `json:"name"`
	Age   *int   `json:"age"`
}
type GenderBody struct {
	Count       int     `json:"count"`
	Name        string  `json:"name"`
	Gender      *string `json:"gender"`
	Probability float64 `json:"probability"`
}

type NationalityBody struct {
	Count       int    `json:"count"`
	Name        string `json:"name"`
	Nationality []struct {
		Country_id  string  `json:"country_id"`
		Probability float64 `json:"probability"`
	} `json:"country"`
}

// Send requests to given api url and writes decoded body
// to result variable
func getData[T any](apiURI *string, result *T) error {
	response, err := http.Get(*apiURI)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(result); err != nil {
		return err
	}
	return nil
}

// Enriches given FIO w/ age, gender and nationality
// by requesting provided API
func enrichData(bodyData *BodyData) (EnrichedData, error) {
	ageAPI := "https://api.agify.io/?name=" + *bodyData.Name
	genderAPI := "https://api.genderize.io/?name=" + *bodyData.Name
	nationalityAPI := "https://api.nationalize.io/?name=" + *bodyData.Name
	var ageRes AgeBody
	var genderRes GenderBody
	var nationalityRes NationalityBody

	if err := getData[AgeBody](&ageAPI, &ageRes); err != nil {
		return EnrichedData{}, err
	}
	if err := getData[GenderBody](&genderAPI, &genderRes); err != nil {
		return EnrichedData{}, err
	}
	if err := getData[NationalityBody](&nationalityAPI, &nationalityRes); err != nil {
		return EnrichedData{}, err
	}

	if ageRes.Age == nil || genderRes.Gender == nil || len(nationalityRes.Nationality) == 0 {
		return EnrichedData{}, errors.New("Provided FIO does not exist")
	}
	return EnrichedData{
		Name:        *bodyData.Name,
		Surname:     *bodyData.Surname,
		Patronymic:  *bodyData.Patronymic,
		Age:         *ageRes.Age,
		Gender:      *genderRes.Gender,
		Nationality: nationalityRes.Nationality[0].Country_id,
	}, nil
}

// Waits for new enriched data in channel
func enrichListener(ctx context.Context, dataChannel <-chan BodyData, dbChannel chan<- EnrichedData, errorsChannel chan<- []byte) {
	for {
		select {
		case data := <-dataChannel:
			enriched, err := enrichData(&data)
			if err != nil {
				errorsChannel <- prepareErrorBytes[BodyData](err.Error(), &data)
				break
			}
			dbChannel <- enriched
		case <-ctx.Done():
			fmt.Printf("Enriched data listener stopped.\n")
			return
		}
	}
}
