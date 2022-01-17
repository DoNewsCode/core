/*
Package container includes the Container type, witch contains a collection of modules.
*/
package container

import (
	"github.com/DoNewsCode/core/contract"
)

var _ contract.Container = (*Container)(nil)

// Container holds all modules registered.
type Container struct {
	modules []interface{}
}

// Modules returns all modules in the container. This method is used to scan for
// custom interfaces. For example, The database module use Modules to scan for
// database migrations.
/*
	m.container.Modules().Filter(func(p MigrationProvider) {
		for _, migration := range p.ProvideMigration() {
			if migration.Connection == "" {
				migration.Connection = "default"
			}
			if migration.Connection == connection {
				migrations.Collection = append(migrations.Collection, migration)
			}
		}
	})
*/
func (c *Container) Modules() []interface{} {
	return c.modules
}

func (c *Container) AddModule(module interface{}) {
	c.modules = append(c.modules, module)
}
