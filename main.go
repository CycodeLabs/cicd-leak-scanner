package main

import "github.com/CycodeLabs/cicd-leak-scanner/cmd"

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
