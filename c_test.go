package core

import (
	"context"
	"github.com/DoNewsCode/core/otgorm"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/DoNewsCode/core/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestC_Serve(t *testing.T) {
	c := New()
	c.ProvideEssentials()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	e := c.Serve(ctx)
	assert.NoError(t, e)
}

func TestC_Default(t *testing.T) {
	c := New()
	c.ProvideEssentials()
	c.Provide(otgorm.Provide)
	c.AddModuleFunc(config.New)

	f, _ := ioutil.TempFile("./", "*")

	rootCommand := &cobra.Command{}
	c.ApplyRootCommand(rootCommand)
	rootCommand.SetArgs([]string{"config", "init", "-o", f.Name()})
	rootCommand.Execute()

	output, _ := ioutil.ReadFile(f.Name())
	assert.Contains(t, string(output), "gorm:")
	os.Remove(f.Name())
}
