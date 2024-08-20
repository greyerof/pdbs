# pdbs: helper tool for the RedHat's Best Practices for K8s CertSuite
This tool reads a claim.json that has been created with [certsuite](https://github.com/redhat-best-practices-for-k8s/certsuite). and creates a CSV with the PodDisruptionBudgets and their targets with this format:
```
pdb-name, pdb-minAvailable, pdb-maxUnavailable, target-type, target-name, target-replicas
```

# How to use it
## Build it
```
go build .
```
## Run it
There's only one flag `--claim`, which should point to your local claim file.
E.g.
```
$ ./pdbs --claim ~/claim.json
```
