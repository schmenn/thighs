package main

import (
	"encoding/json"
	"github.com/fatih/color"
	"os"
)

type PlaceFile struct {
	Places []Place
}

type Place struct {
	ID   string  `json:"id"`
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
	Name string  `json:"name"`
}

func ParsePlacesFile() (*[]Place, error) {
	var placeFile *PlaceFile
	j, err := os.Open(PlacesFile)
	if err != nil {
		color.HiYellow("[!] A places.json file was not detected; this is fine")
		return nil, nil
	}
	color.HiBlack("places.json file detected; parsing")
	err = json.NewDecoder(j).Decode(&placeFile)
	if err != nil {
		return nil, err
	}
	return &placeFile.Places, nil
}
