package generate

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config 配置参数
type Config struct {
	Dir      string
	PkgName  string
	Name     string
	Comment  string
	File     string
	Storage  string
	Modules  string
	Override bool
}

// Exec 执行生成模块命令
func Exec(cfg Config) error {
	cmd := &Command{cfg: &cfg}
	return cmd.Exec()
}

// Command 生成命令
type Command struct {
	cfg *Config
}

func (a *Command) hasModule(m string) bool {
	if v := a.cfg.Modules; v == "" || v == "all" {
		return true
	}

	for _, s := range strings.Split(a.cfg.Modules, ",") {
		if s == m {
			return true
		}
	}

	return false
}

func (a *Command) handleError(err error, desc string) {
	if err != nil {
		fmt.Printf("%s:%s", desc, err.Error())
	}
}

// Exec 执行命令
func (a *Command) Exec() error {
	var item TplItem

	if a.cfg.File != "" {
		b, err := readFile(a.cfg.File)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(b.Bytes(), &item)
		if err != nil {
			return err
		}
	} else {
		item.StructName = a.cfg.Name
		item.Comment = a.cfg.Comment
	}

	dir, err := filepath.Abs(a.cfg.Dir)
	if err != nil {
		return err
	}

	pkgName := a.cfg.PkgName
	ctx := context.Background()
	ctx = context.WithValue(ctx, "override", a.cfg.Override)

	if a.hasModule("schema") {
		err = genSchema(ctx, pkgName, dir, item.StructName, item.Comment, item.toSchemaFields()...)
		if err != nil {
			return err
		}
	}

	if a.hasModule("model") {
		err = genModel(ctx, pkgName, dir, item.StructName, item.Comment)
		a.handleError(err, "Generate model interface")
		switch a.cfg.Storage {
		case "mongo":
			err = genMongoEntity(ctx, pkgName, dir, item.StructName, item.Comment, item.toEntityMongoFields()...)
			a.handleError(err, "Generate mongo entity")

			err = insertEntityInjectMongo(ctx, dir, item.StructName)
			a.handleError(err, "Insert mongo entity inject")

			err = genModelImplMongo(ctx, pkgName, dir, item.StructName, item.Comment)
			a.handleError(err, "Generate mongo model")

			err = insertModelInjectMongo(ctx, dir, item.StructName)
			a.handleError(err, "Insert mongo model inject")
		default:
			err = genGormEntity(ctx, pkgName, dir, item.StructName, item.Comment, item.toEntityGormFields()...)
			a.handleError(err, "Generate gorm entity")

			err = insertEntityInjectGorm(ctx, dir, item.StructName)
			a.handleError(err, "Insert gorm entity inject")

			err = genModelImplGorm(ctx, pkgName, dir, item.StructName, item.Comment)
			a.handleError(err, "Generate gorm model")

			err = insertModelInjectGorm(ctx, dir, item.StructName)
			a.handleError(err, "Insert gorm model inject")
		}
	}

	if a.hasModule("bll") {
		err = genBll(ctx, pkgName, dir, item.StructName, item.Comment)
		a.handleError(err, "Generate bll interface")

		err = genBllImpl(ctx, pkgName, dir, item.StructName, item.Comment)
		a.handleError(err, "Generate bll impl")

		err = insertBllInject(ctx, dir, item.StructName)
		a.handleError(err, "Insert bll inject")
	}

	if a.hasModule("api") {
		err = genAPI(ctx, pkgName, dir, item.StructName, item.Comment)
		a.handleError(err, "Generate api")

		err = insertAPIInject(ctx, dir, item.StructName)
		a.handleError(err, "Insert api inject")
	}

	if a.hasModule("mock") {
		err = genAPIMock(ctx, pkgName, dir, item.StructName, item.Comment)
		a.handleError(err, "Generate api mock")

		err = insertAPIMockInject(ctx, dir, item.StructName)
		a.handleError(err, "Insert api mock inject")
	}

	if a.hasModule("router") {
		err = insertRouterAPI(ctx, dir, item.StructName)
		a.handleError(err, "Insert router api")

		err = insertRouterInject(ctx, dir, item.StructName)
		a.handleError(err, "Insert router inject")
	}

	return nil
}
