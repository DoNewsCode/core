// +build integration

package otgorm_test

import (
	"fmt"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/otgorm"
	"gorm.io/gorm"
)

func Example() {
	// Suppress log output by gorm in this test.
	c := core.New(core.WithInline("log.level", "warn"))
	c.ProvideEssentials()
	c.Provide(otgorm.Provide)
	err := c.Invoke(func(db *gorm.DB) {
		fmt.Println(db.Name())
	})
	fmt.Println(err)
	// Output:
	// mysql
	// <nil>
}
