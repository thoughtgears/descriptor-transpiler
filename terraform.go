package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type TerraformOption func(*TerraformConfiguration) error
type TerraformConfiguration struct {
	Region   string
	Database *TerraformDatabaseConfiguration
}

type TerraformDatabaseConfiguration struct {
	Name    string
	Size    string
	Version string
}

var dbSizeMap = map[Size]string{
	SizeSmall:  "db-g1-small",
	SizeMedium: "db-perf-optimized-N-2",
	SizeLarge:  "db-perf-optimized-N-16",
}

var allowedVersions = map[int]bool{
	15: true,
	16: true,
	17: true,
}

func NewTerraformConfiguration(region string, opts ...TerraformOption) (*TerraformConfiguration, error) {
	config := &TerraformConfiguration{
		Region: region,
	}
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}
	return config, nil
}

// WithDatabase configures the database for the terraform configuration
func WithDatabase(name string, size string, version int) TerraformOption {
	return func(c *TerraformConfiguration) error {
		sizeType := Size(size)
		dbSize, ok := dbSizeMap[sizeType]
		if !ok {
			return fmt.Errorf("invalid database size: %s (must be small, medium, or large)", size)
		}

		if !allowedVersions[version] {
			return errors.New("invalid postgres version: must be 15, 16, or 17")
		}

		c.Database = &TerraformDatabaseConfiguration{
			Name:    fmt.Sprintf("db-%s", name),
			Size:    dbSize,
			Version: fmt.Sprintf("POSTGRES_%d", version),
		}
		return nil
	}
}
func (c *TerraformConfiguration) WriteFile() error {
	filename := "./dist/main.tf"
	hclFile := hclwrite.NewEmptyFile()
	if c.Database != nil {
		c.writeDBModule(hclFile)
	}

	return os.WriteFile(filename, hclFile.Bytes(), 0644)
}

func (c *TerraformConfiguration) writeDBModule(file *hclwrite.File) {
	module := file.Body().AppendNewBlock("module", []string{fmt.Sprintf("database-%s", c.Database.Name)})
	moduleBody := module.Body()
	moduleBody.SetAttributeValue("source", cty.StringVal("./modules/database"))
	moduleBody.AppendNewline()
	moduleBody.SetAttributeValue("name", cty.StringVal(c.Database.Name))
	moduleBody.SetAttributeValue("region", cty.StringVal(c.Region))
	moduleBody.SetAttributeValue("size", cty.StringVal(c.Database.Size))
	moduleBody.SetAttributeValue("db_version", cty.StringVal(c.Database.Version))
}
