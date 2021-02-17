package otgorm

import (
	"github.com/DoNewsCode/std/pkg/core"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModule_ProvideCommand(t *testing.T) {
	c := core.New()
	c.AddCoreDependencies()
	c.AddDependencyFunc(ProvideDatabase)
	c.AddModuleFunc(New)
	rootCmd := cobra.Command{}
	for _, p := range c.GetCommandProviders() {
		p(&rootCmd)
	}
	assert.True(t, rootCmd.HasSubCommands())
}
