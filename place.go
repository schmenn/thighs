package main

import (
	"encoding/json"
	"github.com/fatih/color"
	"os"
)

type Place struct {
	ID   string  `json:"id"`
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
	Name string  `json:"name"`
}

func ParsePlacesFile() (*[]Place, error) {
	var placeFile []Place
	j, err := os.ReadFile(PlacesFile)
	if err != nil {
		color.HiYellow("[!] A places.json file was not detected; this is fine")
		return nil, err
	}
	err = json.Unmarshal(j, &placeFile)
	if err != nil {
		color.HiYellow("[!] places.json could not be read; this is fine")
		return nil, err
	}
	return &placeFile, nil
}
