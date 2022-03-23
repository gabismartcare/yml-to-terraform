package main

import (
	"os"
	"testing"
)

func TestCompiler(t *testing.T) {
	os.Setenv("POSTGRES_PASSWORD", "pg_value")
	os.Setenv("BUILD_VERSION", "version")
	os.Setenv("PROJECT_ID", "projet")
	transformToTerraform("./test/input/test-file.yml", "./test/output")

}
