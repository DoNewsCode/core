package otgorm

import (
	"fmt"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"gorm.io/gorm"
)

// AddGormCallbacks adds callbacks for tracing, you should call SetSpanToGorm to make them work
func AddGormCallbacks(db *gorm.DB, tracer opentracing.Tracer) {
	callbacks := newCallbacks(tracer)
	registerCallbacks(db, "create", callbacks)
	registerCallbacks(db, "query", callbacks)
	registerCallbacks(db, "update", callbacks)
	registerCallbacks(db, "delete", callbacks)
	registerCallbacks(db, "row_query", callbacks)
}

type callbacks struct {
	tracer opentracing.Tracer
}

func newCallbacks(tracer opentracing.Tracer) *callbacks {
	return &callbacks{tracer}
}

func (c *callbacks) beforeCreate(scope *gorm.DB)   { c.before(scope) }
func (c *callbacks) afterCreate(scope *gorm.DB)    { c.after(scope, "INSERT") }
func (c *callbacks) beforeQuery(scope *gorm.DB)    { c.before(scope) }
func (c *callbacks) afterQuery(scope *gorm.DB)     { c.after(scope, "SELECT") }
func (c *callbacks) beforeUpdate(scope *gorm.DB)   { c.before(scope) }
func (c *callbacks) afterUpdate(scope *gorm.DB)    { c.after(scope, "UPDATE") }
func (c *callbacks) beforeDelete(scope *gorm.DB)   { c.before(scope) }
func (c *callbacks) afterDelete(scope *gorm.DB)    { c.after(scope, "DELETE") }
func (c *callbacks) beforeRowQuery(scope *gorm.DB) { c.before(scope) }
func (c *callbacks) afterRowQuery(scope *gorm.DB)  { c.after(scope, "") }

func (c *callbacks) before(db *gorm.DB) {
	span, newCtx := opentracing.StartSpanFromContextWithTracer(db.Statement.Context, c.tracer, "sql")
	ext.DBType.Set(span, "sql")
	db.Statement.WithContext(newCtx)
	db.Set("span", span)
}

func (c *callbacks) after(db *gorm.DB, operation string) {
	spanInterface, ok := db.Get("span")
	if !ok {
		return
	}
	span := spanInterface.(opentracing.Span)
	if operation == "" {
		operation = strings.ToUpper(strings.Split(db.Statement.SQL.String(), " ")[0])
	}
	ext.Error.Set(span, db.Error != nil)
	ext.DBStatement.Set(span, db.Statement.SQL.String())
	span.SetTag("db.table", db.Statement.Table)
	span.SetTag("db.method", operation)
	span.SetTag("db.err", db.Error != nil)
	span.SetTag("db.count", db.Statement.RowsAffected)
	span.Finish()
}

func registerCallbacks(db *gorm.DB, name string, c *callbacks) {
	beforeName := fmt.Sprintf("tracing:%v_before", name)
	afterName := fmt.Sprintf("tracing:%v_after", name)
	gormCallbackName := fmt.Sprintf("gorm:%v", name)
	// gorm does some magic, if you pass CallbackProcessor here - nothing works
	switch name {
	case "create":
		db.Callback().Create().Before(gormCallbackName).Register(beforeName, c.beforeCreate)
		db.Callback().Create().After(gormCallbackName).Register(afterName, c.afterCreate)
	case "query":
		db.Callback().Query().Before(gormCallbackName).Register(beforeName, c.beforeQuery)
		db.Callback().Query().After(gormCallbackName).Register(afterName, c.afterQuery)
	case "update":
		db.Callback().Update().Before(gormCallbackName).Register(beforeName, c.beforeUpdate)
		db.Callback().Update().After(gormCallbackName).Register(afterName, c.afterUpdate)
	case "delete":
		db.Callback().Delete().Before(gormCallbackName).Register(beforeName, c.beforeDelete)
		db.Callback().Delete().After(gormCallbackName).Register(afterName, c.afterDelete)
	case "row_query":
		db.Callback().Row().Before(gormCallbackName).Register(beforeName, c.beforeRowQuery)
		db.Callback().Row().After(gormCallbackName).Register(afterName, c.afterRowQuery)
	}
}
