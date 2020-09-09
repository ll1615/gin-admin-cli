package generate

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gin-admin/gin-admin-cli/util"
)

type entityGormField struct {
	Name        string // 字段名
	Comment     string // 字段注释
	Type        string // 字段类型
	GormOptions string // gorm配置项(不包含column)
}

func getEntityGormFileName(dir, name string) string {
	fullname := fmt.Sprintf("%s/app/model/impl/gorm/entity/e_%s.go", dir, util.ToLowerUnderlinedNamer(name))
	return fullname
}

// 生成entity文件
func genGormEntity(ctx context.Context, pkgName, dir, name, comment string, fields ...entityGormField) error {
	var tfields []entityGormField

	tfields = append(tfields, fields...)

	buf := new(bytes.Buffer)
	for _, field := range tfields {
		buf.WriteString(fmt.Sprintf("%s \t %s \t", field.Name, field.Type))
		buf.WriteByte('`')

		gormTag := fmt.Sprintf("column:%s;", util.ToLowerUnderlinedNamer(field.Name))
		if field.GormOptions != "" {
			gormTag = field.GormOptions
		}
		buf.WriteString(fmt.Sprintf(`gorm:"%s"`, gormTag))

		buf.WriteByte('`')

		if field.Comment != "" {
			buf.WriteString(fmt.Sprintf("// %s", field.Comment))
		}
		buf.WriteString(delimiter)
	}

	tbuf, err := execParseTpl(entityGormTpl, map[string]interface{}{
		"PkgName":       pkgName,
		"Name":          name,
		"PluralName":    util.ToPlural(name),
		"Fields":        buf.String(),
		"Comment":       comment,
		"UnderLineName": util.ToLowerUnderlinedNamer(name),
	})
	if err != nil {
		return err
	}

	filename := getEntityGormFileName(dir, name)
	err = createFile(ctx, filename, tbuf)
	if err != nil {
		return err
	}

	fmt.Printf("文件[%s]写入成功\n", filename)

	return execGoFmt(filename)
}

const entityGormTpl = `
package entity

import (
	"context"

	"{{.PkgName}}/app/schema"
	"{{.PkgName}}/pkg/util"

	"gorm.io/gorm"
)

// Get{{.Name}}DB 获取{{.Name}}存储
func Get{{.Name}}DB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return GetDBWithModel(ctx, defDB, new({{.Name}}))
}

// Schema{{.Name}} {{.Comment}}对象
type Schema{{.Name}} schema.{{.Name}}

// To{{.Name}} 转换为实体
func (a Schema{{.Name}}) To{{.Name}}() *{{.Name}} {
	item := new({{.Name}})
	util.StructMapToStruct(a, item)
	return item
}

// {{.Name}} {{.Comment}}实体
type {{.Name}} struct {
	{{.Fields}}
}

// ToSchema{{.Name}} 转换为demo对象
func (a {{.Name}}) ToSchema{{.Name}}() *schema.{{.Name}} {
	item := new(schema.{{.Name}})
	util.StructMapToStruct(a, item)
	return item
}

// {{.PluralName}} {{.Comment}}实体列表
type {{.PluralName}} []*{{.Name}}

// ToSchema{{.PluralName}} 转换为对象列表
func (a {{.PluralName}}) ToSchema{{.PluralName}}() schema.{{.PluralName}} {
	list := make(schema.{{.PluralName}}, len(a))
	for i, item := range a {
		list[i] = item.ToSchema{{.Name}}()
	}
	return list
}

`
