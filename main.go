package main

import (
	"log"

	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

const specPath = "examples/app.yaml"

func main() {
	appSpec, err := LoadAppSpec(specPath)
	if err != nil {
		log.Fatalf("error loading app specification: %v", err)
	}

	labels := appSpec.BuildLabels()

	config, err := NewChartConfiguration(appSpec.Name, "v1.0.0", WithLabels(labels), WithSize(appSpec.ToChartSize()))
	if err != nil {
		log.Fatalf("error generating chart: %v", err)
	}

	app := cdk8s.NewApp(nil)

	config.NewCore(app, config.Name)
	config.Render(app)
}
