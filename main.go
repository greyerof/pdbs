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
	claimFileFlag *string = flag.String("claim", "", "certsuite's claim.json")
)

func init() {
	flag.Parse()
}

var (
	claim = CertsuiteClaim{}
)

type ReplicaSet struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		Replicas int `json:"replicas"`
		Template struct {
			Metadata struct {
				Labels map[string]string `json:"labels"`
			} `json:"metadata"`
		} `json:"template"`
	} `json:"spec"`
}

type CertsuiteClaim struct {
	Claim struct {
		Configurations struct {
			PodDisruptionBudgets []struct {
				Metadata struct {
					Name string `json:"name"`
				} `json:"metadata"`
				Spec struct {
					MinAvailable   *intstr.IntOrString
					MaxUnavailable *intstr.IntOrString
					Selector       struct {
						MatchLabels map[string]string `json:"matchLabels"`
					} `json:"selector"`
				} `json:"spec"`
			} `json:"PodDisruptionBudgets"`

			TestDeployments  []ReplicaSet `json:"testDeployments"`
			TestStatefulSets []ReplicaSet `json:"testStatefulSets"`
		} `json:"configurations"`
	} `json:"claim"`
}

func parseJSONs() {
	if *claimFileFlag == "" {
		fmt.Println("--claim flag cannot be empty. Please provide a valid path for the claim file.")
		os.Exit(1)
	}

	claimFile, err := os.ReadFile(*claimFileFlag)
	if err != nil {
		panic("Failed to read claim file: " + err.Error())
	}

	err = json.Unmarshal(claimFile, &claim)
	if err != nil {
		panic("Failed to unmarshal claim file: " + err.Error())
	}
}

func main() {
	parseJSONs()

	// For each PDB, check show pdb's minAvailable, maxUnavailable and target's name, type and replicas
	// Print CSV HEADER:
	fmt.Println("pdb-name, pdb-minAvailable, pdb-maxUnavailable, target-type, target-name, target-replicas")

	for _, pdb := range claim.Claim.Configurations.PodDisruptionBudgets {
		pdbLabelsSelector := metav1.LabelSelector{MatchLabels: pdb.Spec.Selector.MatchLabels}
		pdbSelector, err := metav1.LabelSelectorAsSelector(&pdbLabelsSelector)
		if err != nil {
			panic("Failed to create pdb selector for sts: " + err.Error())
		}

		targetFound := false
		// Search target:
		for _, dep := range claim.Claim.Configurations.TestDeployments {
			depLabels := labels.Set(dep.Spec.Template.Metadata.Labels)

			if pdbSelector.Matches(depLabels) {
				fmt.Println(pdb.Metadata.Name, ",", pdb.Spec.MinAvailable.String(), ",", pdb.Spec.MaxUnavailable.String(), ",", "deployment", ",", dep.Metadata.Name, ",", dep.Spec.Replicas)
				targetFound = true
			}
		}

		for _, sts := range claim.Claim.Configurations.TestStatefulSets {
			stsLabels := labels.Set(sts.Spec.Template.Metadata.Labels)

			if pdbSelector.Matches(stsLabels) {
				if targetFound {
					panic("target already found in deps for pdb " + pdb.Metadata.Name)
				}

				fmt.Println(pdb.Metadata.Name, ",", pdb.Spec.MinAvailable.String(), ",", pdb.Spec.MaxUnavailable.String(), ",", "statefulset", ",", sts.Metadata.Name, ",", sts.Spec.Replicas)
				targetFound = true
			}
		}

		if !targetFound {
			fmt.Println(pdb.Metadata.Name, ",", pdb.Spec.MinAvailable.String(), ",", pdb.Spec.MaxUnavailable.String(), ",", "TARGET-NOT-FOUND", ",", "TARGET-NOT-FOUND", ",", -1)
		}
	}
}
