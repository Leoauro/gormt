package genfunc

const (
	genTnf = `
// TableName get sql table name.获取数据库表名
func (m *{{.StructName}}) TableName() string {
	return "{{.TableName}}"
}
`
	genColumn = `
// {{.StructName}}Columns get sql column name.获取数据库列名
var {{.StructName}}Columns = struct { {{range $em := .Em}}
	{{$em.StructName}} string{{end}}    
	}{ {{range $em := .Em}}
		{{$em.StructName}}:"{{$em.ColumnName}}",  {{end}}           
	}
`
	genBase = `
package {{.PackageName}}
import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

var globalIsRelated bool = true  // 全局预加载

// prepare for other
type BaseMgr struct {
	*gorm.DB
	ctx       context.Context
	cancel    context.CancelFunc
	timeout   time.Duration
	isRelated bool
}

// SetCtx set context
func (obj *BaseMgr) SetTimeOut(timeout time.Duration) {
	obj.ctx, obj.cancel = context.WithTimeout(context.Background(), timeout)
	obj.timeout = timeout
}

// SetCtx set context
func (obj *BaseMgr) SetCtx(c context.Context) {
	if c != nil {
		obj.ctx = c
	}
}

// Ctx get context
func (obj *BaseMgr) GetCtx() context.Context {
	return obj.ctx
}

// Cancel cancel context
func (obj *BaseMgr) Cancel(c context.Context) {
	obj.cancel()
}

// GetDB get gorm.DB info
func (obj *BaseMgr) GetDB() *gorm.DB {
	return obj.DB
}

// UpdateDB update gorm.DB info
func (obj *BaseMgr) UpdateDB(db *gorm.DB) {
	obj.DB = db
}

// GetIsRelated Query foreign key Association.获取是否查询外键关联(gorm.Related)
func (obj *BaseMgr) GetIsRelated() bool {
	return obj.isRelated
}

// SetIsRelated Query foreign key Association.设置是否查询外键关联(gorm.Related)
func (obj *BaseMgr) SetIsRelated(b bool) {
	obj.isRelated = b
}

// New new gorm.新gorm
func (obj *BaseMgr) New() *gorm.DB {
	return obj.DB.Session(&gorm.Session{ Context: obj.ctx})
}

type options struct {
	query map[string]interface{}
}

// Option overrides behavior of Connect.
type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}


// OpenRelated 打开全局预加载
func OpenRelated() {
	globalIsRelated = true
}

// CloseRelated 关闭全局预加载
func CloseRelated() {
	globalIsRelated = true
}


// 自定义sql查询
type Condetion struct {
	list []*condetionInfo
}

// And a condition by and .and 一个条件
func (c *Condetion) And(column string, cases string, value ...interface{}) {
	c.list = append(c.list, &condetionInfo{
		andor:  "and",
		column: column, // 列名
		case_:  cases,  // 条件(and,or,in,>=,<=)
		value:  value,
	})
}

// Or a condition by or .or 一个条件
func (c *Condetion) Or(column string, cases string, value ...interface{}) {
	c.list = append(c.list, &condetionInfo{
		andor:  "or",
		column: column, // 列名
		case_:  cases,  // 条件(and,or,in,>=,<=)
		value:  value,
	})
}

func (c *Condetion) Get() (where string, out []interface{}) {
	firstAnd := -1
	for i := 0; i < len(c.list); i++ { // 查找第一个and
		if c.list[i].andor == "and" {
			where = fmt.Sprintf("{{GetVV }} %v ?", c.list[i].column, c.list[i].case_)
			out = append(out, c.list[i].value)
			firstAnd = i
			break
		}
	}

	if firstAnd < 0 && len(c.list) > 0 { // 补刀
		where = fmt.Sprintf("{{GetVV }} %v ?", c.list[0].column, c.list[0].case_)
		out = append(out, c.list[0].value)
		firstAnd = 0
	}

	for i := 0; i < len(c.list); i++ { // 添加剩余的
		if firstAnd != i {
			where += fmt.Sprintf(" %v {{GetVV }} %v ?", c.list[i].andor, c.list[i].column, c.list[i].case_)
			out = append(out, c.list[i].value)
		}
	}

	return
}

type condetionInfo struct {
	andor  string
	column string // 列名
	case_  string // 条件(in,>=,<=)
	value  interface{}
}
	`

	genlogic = `{{$obj := .}}{{$list := $obj.Em}}
type {{$obj.StructName}}Mgr struct {
	*BaseMgr
}

// {{$obj.StructName}}Mgr open func
func New{{$obj.StructName}}Mgr(db *gorm.DB) *{{$obj.StructName}}Mgr {
	if db == nil {
		panic(fmt.Errorf("{{$obj.StructName}}Mgr need init by db"))
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &{{$obj.StructName}}Mgr{BaseMgr: &BaseMgr{DB: db.Model({{$obj.StructName}}{}), isRelated: globalIsRelated,ctx:ctx,cancel:cancel,timeout:-1}}
}

// GetTableName get sql table name.获取数据库名字
func (obj *{{$obj.StructName}}Mgr) GetTableName() string {
	return "{{GetTablePrefixName $obj.TableName}}"
}

// Get 获取
func (obj *{{$obj.StructName}}Mgr) Get() (result {{$obj.StructName}}, err error) {
	err = obj.DB.WithContext(obj.ctx).Table(obj.GetTableName()).Find(&result).Error
	{{GenPreloadList $obj.PreloadList false}}
	return
}

// Gets 获取批量结果
func (obj *{{$obj.StructName}}Mgr) Gets() (results []*{{$obj.StructName}}, err error) {
	err = obj.DB.WithContext(obj.ctx).Table(obj.GetTableName()).Find(&results).Error
	{{GenPreloadList $obj.PreloadList true}}
	return
}

//////////////////////////option case ////////////////////////////////////////////
{{range $oem := $obj.Em}}
// With{{$oem.ColStructName}} {{$oem.ColName}}获取 {{$oem.Notes}}
func (obj *{{$obj.StructName}}Mgr) With{{$oem.ColStructName}}({{CapLowercase $oem.ColStructName}} {{$oem.Type}}) Option {
	return optionFunc(func(o *options) { o.query["{{$oem.ColName}}"] = {{CapLowercase $oem.ColStructName}} })
}
{{end}}

// GetByOption 功能选项模式获取
func (obj *{{$obj.StructName}}Mgr) GetByOption(opts ...Option) (result {{$obj.StructName}}, err error) {
	options := options{
		query: make(map[string]interface{}, len(opts)),
	}
	for _, o := range opts {
		o.apply(&options)
	}

	err = obj.DB.WithContext(obj.ctx).Table(obj.GetTableName()).Where(options.query).Find(&result).Error
	{{GenPreloadList $obj.PreloadList false}}
	return
}

// GetByOptions 批量功能选项模式获取
func (obj *{{$obj.StructName}}Mgr) GetByOptions(opts ...Option) (results []*{{$obj.StructName}}, err error) {
	options := options{
		query: make(map[string]interface{}, len(opts)),
	}
	for _, o := range opts {
		o.apply(&options)
	}

	err = obj.DB.WithContext(obj.ctx).Table(obj.GetTableName()).Where(options.query).Find(&results).Error
	{{GenPreloadList $obj.PreloadList true}}
	return
}
//////////////////////////enume case ////////////////////////////////////////////

{{range $oem := $obj.Em}}
// GetFrom{{$oem.ColStructName}} 通过{{$oem.ColName}}获取内容 {{$oem.Notes}} {{if $oem.IsMulti}}
func (obj *{{$obj.StructName}}Mgr) GetFrom{{$oem.ColStructName}}({{CapLowercase $oem.ColStructName}} {{$oem.Type}}) (results []*{{$obj.StructName}}, err error) {
	err = obj.DB.WithContext(obj.ctx).Table(obj.GetTableName()).Where("{{$oem.ColNameEx}} = ?", {{CapLowercase $oem.ColStructName}}).Find(&results).Error
	{{GenPreloadList $obj.PreloadList true}}
	return
}
{{else}}
func (obj *{{$obj.StructName}}Mgr)  GetFrom{{$oem.ColStructName}}({{CapLowercase $oem.ColStructName}} {{$oem.Type}}) (result {{$obj.StructName}}, err error) {
	err = obj.DB.WithContext(obj.ctx).Table(obj.GetTableName()).Where("{{$oem.ColNameEx}} = ?", {{CapLowercase $oem.ColStructName}}).Find(&result).Error
	{{GenPreloadList $obj.PreloadList false}}
	return
}
{{end}}
// GetBatchFrom{{$oem.ColStructName}} 批量查找 {{$oem.Notes}}
func (obj *{{$obj.StructName}}Mgr) GetBatchFrom{{$oem.ColStructName}}({{CapLowercase $oem.ColStructName}}s []{{$oem.Type}}) (results []*{{$obj.StructName}}, err error) {
	err = obj.DB.WithContext(obj.ctx).Table(obj.GetTableName()).Where("{{$oem.ColNameEx}} IN (?)", {{CapLowercase $oem.ColStructName}}s).Find(&results).Error
	{{GenPreloadList $obj.PreloadList true}}
	return
}
 {{end}}
 //////////////////////////primary index case ////////////////////////////////////////////
 {{range $ofm := $obj.Primary}}
 // {{GenFListIndex $ofm 1}} primary or index 获取唯一内容
 func (obj *{{$obj.StructName}}Mgr) {{GenFListIndex $ofm 1}}({{GenFListIndex $ofm 2}}) (result {{$obj.StructName}}, err error) {
	err = obj.DB.WithContext(obj.ctx).Table(obj.GetTableName()).Where("{{GenFListIndex $ofm 3}}", {{GenFListIndex $ofm 4}}).Find(&result).Error
	{{GenPreloadList $obj.PreloadList false}}
	return
}
 {{end}}

 {{range $ofm := $obj.Index}}
 // {{GenFListIndex $ofm 1}}  获取多个内容
 func (obj *{{$obj.StructName}}Mgr) {{GenFListIndex $ofm 1}}({{GenFListIndex $ofm 2}}) (results []*{{$obj.StructName}}, err error) {
	err = obj.DB.WithContext(obj.ctx).Table(obj.GetTableName()).Where("{{GenFListIndex $ofm 3}}", {{GenFListIndex $ofm 4}}).Find(&results).Error
	{{GenPreloadList $obj.PreloadList true}}
	return
}
 {{end}}

`
	genPreload = `if err == nil && obj.isRelated { {{range $obj := .}}{{if $obj.IsMulti}}
		if err = obj.New().Table("{{$obj.ForeignkeyTableName}}").Where("{{$obj.ForeignkeyCol}} = ?", result.{{$obj.ColStructName}}).Find(&result.{{$obj.ForeignkeyStructName}}List).Error;err != nil { // {{$obj.Notes}}
				if err != gorm.ErrRecordNotFound { // 非 没找到
					return
				}	
			} {{else}} 
		if err = obj.New().Table("{{$obj.ForeignkeyTableName}}").Where("{{$obj.ForeignkeyCol}} = ?", result.{{$obj.ColStructName}}).Find(&result.{{$obj.ForeignkeyStructName}}).Error; err != nil { // {{$obj.Notes}} 
				if err != gorm.ErrRecordNotFound { // 非 没找到
					return
				}
			}{{end}} {{end}}}
`
	genPreloadMulti = `if err == nil && obj.isRelated {
		for i := 0; i < len(results); i++ { {{range $obj := .}}{{if $obj.IsMulti}}
		if err = obj.New().Table("{{$obj.ForeignkeyTableName}}").Where("{{$obj.ForeignkeyCol}} = ?", results[i].{{$obj.ColStructName}}).Find(&results[i].{{$obj.ForeignkeyStructName}}List).Error;err != nil { // {{$obj.Notes}}
				if err != gorm.ErrRecordNotFound { // 非 没找到
					return
				}
			} {{else}} 
		if err = obj.New().Table("{{$obj.ForeignkeyTableName}}").Where("{{$obj.ForeignkeyCol}} = ?", results[i].{{$obj.ColStructName}}).Find(&results[i].{{$obj.ForeignkeyStructName}}).Error; err != nil { // {{$obj.Notes}} 
				if err != gorm.ErrRecordNotFound { // 非 没找到
					return
				}
			} {{end}} {{end}}
	}
}`
)
