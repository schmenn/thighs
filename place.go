package main

import (
	"encoding/json"
	"os"
)

type Place struct {
	ID   string  `json:"id"`
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
	Name string  `json:"name"`
}

func ParsePlacesFile() (*[]Place, error) {
	// checks to make sure file exists first.
	// any errors thrown by stat will mean that the file probably can't be read
	// due to not existing, lack of permissions, etc.
	_, err := os.Stat(PlacesFile)
	if err != nil {
		return nil, nil
	}

	var placeFile []Place
	j, err := os.ReadFile(PlacesFile)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(j, &placeFile)
	if err != nil {
		return nil, err
	}
	return &placeFile, nil
}
