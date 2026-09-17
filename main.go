package main

import (
	"log"

	cli "fizz-buzz-rest/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}
