package main

import (
	"os"
	"testing"
)

func TestCompiler(t *testing.T) {
	os.Setenv("POSTGRES_PASSWORD", "pg_value")
	os.Setenv("BUILD_VERSION", "version")
	os.Setenv("PROJECT_ID", "projet")
	transformToTerraform("/home/tom/IdeaProjects/gabi/integration/pediarity/pediarity.yml", "./test/output")

}
