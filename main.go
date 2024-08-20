package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	Name           string              `json:"name"`
	MinAvailable   *intstr.IntOrString `json:"minAvailable"`
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable"`
	Target         PDBTarget           `json:"target"`
}

type ReplicaSet struct {
	Name     string            `json:"name"`
	Replicas int               `json:"replicas"`
	Labels   map[string]string `json:"labels"`
}

func init() {
	flag.Parse()
}

var (
	pdbs = []PDB{}
	deps = []ReplicaSet{}
	stss = []ReplicaSet{}
)

func parseJSONs() {
	pdbsFile, err := os.ReadFile(*pdbsFlag)
	if err != nil {
		panic("Failed to read pdbs file: " + err.Error())
	}

	err = json.Unmarshal(pdbsFile, &pdbs)
	if err != nil {
		panic("Failed to unmarshall pdbs: " + err.Error())
	}

	depsFile, err := os.ReadFile(*depsFlag)
	if err != nil {
		panic("Failed to read deployments file: " + err.Error())
	}

	err = json.Unmarshal(depsFile, &deps)
	if err != nil {
		panic("Failed to unmarshall deployments: " + err.Error())
	}

	stssFile, err := os.ReadFile(*stssFlag)
	if err != nil {
		panic("Failed to read statefulsets file: " + err.Error())
	}

	err = json.Unmarshal(stssFile, &stss)
	if err != nil {
		panic("Failed to unmarshall statefulsets: " + err.Error())
	}
}

func main() {
	parseJSONs()

	// logrus.Infof("Pod Disruption Budgets:")
	// for _, pdb := range pdbs {
	// 	logrus.Infof("PDB Name: %v, MinAvailable: %v, MaxUnavailable: %v, Target: %+v", pdb.Name, pdb.MinAvailable, pdb.MaxUnavailable.String(), pdb.Target)
	// }

	// logrus.Infof("Deployments:")
	// for _, dep := range deps {
	// 	logrus.Infof("DEP Name: %v, Replicas: %v, Labels: %+v", dep.Name, dep.Replicas, dep.Labels)
	// }

	// logrus.Infof("StatefulSets:")
	// for _, sts := range stss {
	// 	logrus.Infof("STS Name: %v, Replicas: %v, Labels: %+v", sts.Name, sts.Replicas, sts.Labels)
	// }

	// For each PDB, check show pdb's minAvailable, maxUnavailable and target's name, type and replicas
	// Print CSV HEADER:
	fmt.Println("pdb-name, pdb-minAvailable, pdb-maxUnavailable, target-type, target-name, target-replicas")

	for _, pdb := range pdbs {
		pdbLabelsSelector := metav1.LabelSelector{MatchLabels: pdb.Target.MatchLabels}
		pdbSelector, err := metav1.LabelSelectorAsSelector(&pdbLabelsSelector)
		if err != nil {
			panic("Failed to create pdb selector for sts: " + err.Error())
		}

		targetFound := false
		// Search target:
		for _, dep := range deps {
			depLabels := labels.Set(dep.Labels)

			if pdbSelector.Matches(depLabels) {
				fmt.Println(pdb.Name, ",", pdb.MinAvailable.String(), ",", pdb.MaxUnavailable.String(), ",", "deployment", ",", dep.Name, ",", dep.Replicas)
				targetFound = true
			}
		}

		for _, sts := range stss {
			stsLabels := labels.Set(sts.Labels)

			if pdbSelector.Matches(stsLabels) {
				if targetFound {
					panic("target already found in deps for pdb " + pdb.Name)
				}

				fmt.Println(pdb.Name, ",", pdb.MinAvailable.String(), ",", pdb.MaxUnavailable.String(), ",", "statefulset", ",", sts.Name, ",", sts.Replicas)
				targetFound = true
			}
		}

		if !targetFound {
			fmt.Println(pdb.Name, ",", pdb.MinAvailable.String(), ",", pdb.MaxUnavailable.String(), ",", "TARGET-NOT-FOUND", ",", "TARGET-NOT-FOUND", ",", -1)
		}
	}
}
