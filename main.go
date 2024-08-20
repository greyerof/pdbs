package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

type CertsuiteClaim struct {
	Claim struct {
		Configurations struct {
			PodDisruptionBudgets []*policyv1.PodDisruptionBudget `json:"PodDisruptionBudgets"`

			TestDeployments  []*appsv1.Deployment  `json:"testDeployments"`
			TestStatefulSets []*appsv1.StatefulSet `json:"testStatefulSets"`
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

	// Print CSV HEADER:
	fmt.Println("pdb-name, pdb-minAvailable, pdb-maxUnavailable, target-type, target-name, target-replicas")

	for _, pdb := range claim.Claim.Configurations.PodDisruptionBudgets {
		pdbSelector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
		if err != nil {
			panic("Failed to create pdb selector for sts: " + err.Error())
		}

		targetFound := false
		// Search target:
		for _, dep := range claim.Claim.Configurations.TestDeployments {
			depLabels := labels.Set(dep.Spec.Template.Labels)

			if pdbSelector.Matches(depLabels) {
				fmt.Println(pdb.Name, ",", pdb.Spec.MinAvailable.String(), ",", pdb.Spec.MaxUnavailable.String(), ",", "deployment", ",", dep.Name, ",", *dep.Spec.Replicas)
				targetFound = true
			}
		}

		for _, sts := range claim.Claim.Configurations.TestStatefulSets {
			stsLabels := labels.Set(sts.Spec.Template.Labels)

			if pdbSelector.Matches(stsLabels) {
				if targetFound {
					panic("target already found in deps for pdb " + pdb.Name)
				}

				fmt.Println(pdb.Name, ",", pdb.Spec.MinAvailable.String(), ",", pdb.Spec.MaxUnavailable.String(), ",", "statefulset", ",", sts.Name, ",", *sts.Spec.Replicas)
				targetFound = true
			}
		}

		if !targetFound {
			fmt.Println(pdb.Name, ",", pdb.Spec.MinAvailable.String(), ",", pdb.Spec.MaxUnavailable.String(), ",", "TARGET-NOT-FOUND", ",", "TARGET-NOT-FOUND", ",", -1)
		}
	}
}
