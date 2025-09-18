package main

import (
	"context"
	"flag"
	"log"

	"github.com/disc/terraform-provider-pritunl/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)


//go:generate terraform fmt -recursive ./examples/

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	version string = "dev"

)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	address := "registry.terraform.io/disc/pritunl"

	if debugMode {
		err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
			Address:         address,
			Debug:           true,
			ProtocolVersion: 6,
		})
		if err != nil {
			log.Fatal(err.Error())
		}
		return
	}

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address:         address,
		ProtocolVersion: 6,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}
