package main

import (
	"errors"
	"fmt"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"

	"github.com/thoughtgears/descriptor-transpiler/imports/k8s"
)

// Size represents predefined resource configurations.
type Size string

const (
	SizeTiny   Size = "tiny"
	SizeSmall  Size = "small"
	SizeMedium Size = "medium"
	SizeLarge  Size = "large"
)

// ResourceSpec defines CPU and memory allocations.
type ResourceSpec struct {
	CPU    string
	Memory string
}

var sizeConfigs = map[Size]ResourceSpec{
	SizeTiny:   {CPU: "100m", Memory: "128Mi"},
	SizeSmall:  {CPU: "200m", Memory: "256Mi"},
	SizeMedium: {CPU: "500m", Memory: "512Mi"},
	SizeLarge:  {CPU: "1", Memory: "1Gi"},
}

// ChartConfiguration holds the configuration for a Kubernetes deployment chart.
// It contains all necessary settings to generate a deployment, including container
// specifications, resource limits, and environment configuration.
type ChartConfiguration struct {
	Labels    map[string]string
	Port      int
	Replicas  int
	Image     string
	Name      string
	Resources map[string]string
	CPU       string
	Memory    string
	EnvVars   map[string]string
}

// ChartOption is a functional option for configuring ChartConfiguration.
type ChartOption func(*ChartConfiguration)

// WithPort sets the container port for the deployment.
func WithPort(port int) ChartOption {
	return func(c *ChartConfiguration) {
		c.Port = port
	}
}

// WithReplicas sets the number of pod replicas.
func WithReplicas(replicas int) ChartOption {
	return func(c *ChartConfiguration) {
		c.Replicas = replicas
	}
}

// WithSize sets the resource allocation based on predefined size tiers.
// Valid sizes: tiny, small, medium, large. Defaults to tiny if invalid.
func WithSize(size Size) ChartOption {
	return func(c *ChartConfiguration) {
		spec, ok := sizeConfigs[size]
		if !ok {
			spec = sizeConfigs[SizeTiny]
		}
		c.CPU = spec.CPU
		c.Memory = spec.Memory
	}
}

// WithCustomResources sets custom CPU and memory limits.
func WithCustomResources(cpu, memory string) ChartOption {
	return func(c *ChartConfiguration) {
		c.CPU = cpu
		c.Memory = memory
	}
}

// WithLabels sets the Kubernetes labels for pod metadata.
func WithLabels(labels map[string]string) ChartOption {
	return func(c *ChartConfiguration) {
		c.Labels = labels
	}
}

// WithEnvVars sets environment variables for the container.
func WithEnvVars(envVars map[string]string) ChartOption {
	return func(c *ChartConfiguration) {
		c.EnvVars = envVars
	}
}

// NewChartConfiguration creates a new ChartConfiguration with defaults and applies optional configurations.
// The image is automatically constructed using the GCR registry path with the provided name and tag.
//
// Default values:
//   - Port: 8080
//   - Replicas: 1
//   - CPU: "100m"
//   - Memory: "128Mi"
//
// Returns an error if the name or tag is empty.
func NewChartConfiguration(name, tag string, opts ...ChartOption) (*ChartConfiguration, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	if tag == "" {
		return nil, errors.New("tag cannot be empty")
	}

	c := &ChartConfiguration{
		Name:     name,
		Image:    fmt.Sprintf("europe-docker.pkg.dev/my-gcr-project-1234/apps/%s:%s", name, tag),
		Port:     8080,
		Replicas: 1,
		CPU:      sizeConfigs[SizeTiny].CPU,
		Memory:   sizeConfigs[SizeTiny].Memory,
		Labels:   make(map[string]string),
		EnvVars:  make(map[string]string),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// NewCore generates a CDK8s chart with a Kubernetes deployment.
func (c *ChartConfiguration) NewCore(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), nil)

	k8s.NewKubeDeployment(chart, jsii.String("deployment"), &k8s.KubeDeploymentProps{
		Spec: &k8s.DeploymentSpec{
			Selector: &k8s.LabelSelector{
				MatchLabels: c.convertLabels(),
			},
			Replicas: jsii.Number(c.Replicas),
			Template: &k8s.PodTemplateSpec{
				Metadata: &k8s.ObjectMeta{Labels: c.convertLabels()},
				Spec: &k8s.PodSpec{
					Containers: &[]*k8s.Container{c.buildContainer()},
				},
			},
		},
	})

	return chart
}

func (c *ChartConfiguration) convertLabels() *map[string]*string {
	labels := make(map[string]*string, len(c.Labels))
	for key, value := range c.Labels {
		labels[key] = jsii.String(value)
	}
	return &labels
}

func (c *ChartConfiguration) convertEnvVars() *[]*k8s.EnvVar {
	envVars := make([]*k8s.EnvVar, 0, len(c.EnvVars))
	for key, value := range c.EnvVars {
		envVars = append(envVars, &k8s.EnvVar{
			Name:  jsii.String(key),
			Value: jsii.String(value),
		})
	}
	return &envVars
}

func (c *ChartConfiguration) buildContainer() *k8s.Container {
	return &k8s.Container{
		Name:  jsii.String(c.Name),
		Image: jsii.String(c.Image),
		Resources: &k8s.ResourceRequirements{
			Limits: &map[string]k8s.Quantity{
				"cpu":    k8s.Quantity_FromString(jsii.String(c.CPU)),
				"memory": k8s.Quantity_FromString(jsii.String(c.Memory)),
			},
		},
		Env: c.convertEnvVars(),
		Ports: &[]*k8s.ContainerPort{{
			ContainerPort: jsii.Number(c.Port),
		}},
	}
}

func (c *ChartConfiguration) Render(app cdk8s.App) {
	app.Synth()
}
