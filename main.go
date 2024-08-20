package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	pdbsFlag *string = flag.String("pdbs", "", "json file for pdbs")
	depsFlag *string = flag.String("deps", "", "json file for deployments")
	stssFlag *string = flag.String("stss", "", "json file for statefulsets")
)

type PDBTarget struct {
	MatchLabels map[string]string `json:"matchLabels"`
}

type PDB struct {
	Name           string    `json:"name"`
	MaxUnavailable string    `json:"maxUnavailable"`
	Target         PDBTarget `json:"target"`
}

type Deployment struct {
	Name     string            `json:"name"`
	Replicas int               `json:"replicas"`
	Labels   map[string]string `json:"labels"`
}

func main() {
	pdbsFile, err := os.ReadFile(*pdbsFlag)
	if err != nil {
		panic("Failed to read pdbs file: " + err.Error())
	}

	pdbs := []PDB{}

	err = json.Unmarshal(pdbsFile, &pdbs)
	if err != nil {
		panic("Failed to unmarshall pdbs: " + err.Error())
	}

	for _, pdb := range pdbs {
		logrus.Infof("PDB Name: %v, MaxUnavaialble: %v, Target: %+v", pdb.Name, pdb.MaxUnavailable, pdb.Target)
	}
}
