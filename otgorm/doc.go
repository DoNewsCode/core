/*
Package otgorm provides gorm with opentracing.
For documentation about gorm usage, see https://gorm.io/index.html

Integration

package otgorm exports the configuration in the following format:

	gorm:
	  default:
		database: mysql
		dsn: root@tcp(127.0.0.1:3306)/app?charset=utf8mb4&parseTime=True&loc=Local

Add the gorm dependency to core:

	var c *core.C = core.New()
	c.Provide(otgorm.Providers())

Then you can invoke gorm from the application.

	c.Invoke(func(client *gorm.DB) {
		// use client
	})

The database types you can specify in configs are "mysql", "sqlite" and
"clickhouse". Other database types can be added by injecting the otgorm.Drivers
type to the dependency graph.

For example, if we want to use postgres:

    var c *core.C = core.New()
    c.Provide(otgorm.Providers())
    c.Provide(di.Deps{func() otgorm.Drivers {
    	return otgorm.Drivers{
    	    "mysql":      mysql.Open,
    	    "sqlite":     sqlite.Open,
    	    "clickhouse": clickhouse.Open,
    	    "postgres":   postgres.Open,
        }
    }}

Sometimes there are valid reasons to connect to more than one mysql server.
Inject otgorm.Maker to factory a *gorm.DB with a specific configuration entry.

	c.Invoke(function(maker otgorm.Maker) {
		client, err := maker.Make("default")
		// do something with client
	})

Migration and Seeding

package otgorm comes with migration and seeding support. Other modules can
register migration and seeding that are to be run by the command included in
this package.

To invoke the command, add the module to core first:

	c.AddModuleFunc(otgorm.New)

Then you can migrate the database by running:

	go run main.go database migrate

See examples to learn more.
*/
package otgorm
