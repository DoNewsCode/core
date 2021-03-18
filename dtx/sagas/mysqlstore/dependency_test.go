package mysqlstore

import (
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/dtx/sagas"
	"github.com/DoNewsCode/core/otgorm"
)

func TestProviders(t *testing.T) {
	c := core.Default()
	c.Provide(otgorm.Providers())
	c.Provide(sagas.Providers())
	c.Provide(Providers())
	c.Invoke(func(r *sagas.Registry) {
		if _, ok := r.Store.(*MySQLStore); !ok {
			t.Fatal("r.Store should be a mysql store")
		}
	})
}
