package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/DoNewsCode/core/codec/json"
	"github.com/DoNewsCode/core/codec/yaml"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/contract/lifecycle"
	"github.com/DoNewsCode/core/di"

	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Module is the configuration module that bundles the reload watcher and exportConfig commands.
// This module triggers ReloadedEvent on configuration change.
type Module struct {
	conf            *KoanfAdapter
	dispatcher      lifecycle.ConfigReload
	exportedConfigs []ExportedConfig
}

// ConfigIn is the injection parameter for config.New.
type ConfigIn struct {
	di.In

	Conf            contract.ConfigAccessor
	Dispatcher      lifecycle.ConfigReload `optional:"true"`
	ExportedConfigs []ExportedConfig       `group:"config"`
}

// New creates a new config module. It contains the init command.
func New(p ConfigIn) (Module, error) {
	var (
		ok      bool
		adapter *KoanfAdapter
	)
	if adapter, ok = p.Conf.(*KoanfAdapter); !ok {
		return Module{}, fmt.Errorf("expects a *config.KoanfAdapter instance, but %T given", p.Conf)
	}

	if err := loadValidators(adapter, p.ExportedConfigs); err != nil {
		return Module{}, err
	}

	return Module{
		dispatcher:      p.Dispatcher,
		conf:            adapter,
		exportedConfigs: p.ExportedConfigs,
	}, nil
}

// ProvideRunGroup runs the configuration watcher.
func (m Module) ProvideRunGroup(group *run.Group) {
	ctx, cancel := context.WithCancel(context.Background())
	group.Add(func() error {
		m.conf.dispatcher = m.dispatcher
		return m.conf.Watch(ctx)
	}, func(err error) {
		cancel()
	})
}

// ProvideCommand provides the config related command.
func (m Module) ProvideCommand(command *cobra.Command) {
	var (
		targetFilePath string
		style          string
	)
	initCmd := &cobra.Command{
		Use:   "init [module]",
		Short: "export a copy of default config.",
		Long:  "export a default config for currently installed modules.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				handler         handler
				targetFile      *os.File
				exportedConfigs []ExportedConfig
				confMap         map[string]any
				err             error
			)
			handler, err = getHandler(style)
			if err != nil {
				return err
			}
			if len(args) == 0 {
				exportedConfigs = m.exportedConfigs
			}
			if len(args) >= 1 {
				copy := make([]ExportedConfig, 0)
				for i := range m.exportedConfigs {
					for j := 0; j < len(args); j++ {
						if args[j] == m.exportedConfigs[i].Owner {
							copy = append(copy, m.exportedConfigs[i])
							break
						}
					}
				}
				exportedConfigs = copy
			}
			os.MkdirAll(filepath.Dir(targetFilePath), os.ModePerm)
			targetFile, err = os.OpenFile(targetFilePath,
				handler.flags(), os.ModePerm)
			if err != nil {
				return errors.Wrap(err, "failed to open config file")
			}
			defer targetFile.Close()
			bytes, err := ioutil.ReadAll(targetFile)
			if err != nil {
				return errors.Wrap(err, "failed to read config file")
			}
			err = handler.unmarshal(bytes, &confMap)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal config file")
			}
			err = handler.write(targetFile, exportedConfigs, confMap)
			if err != nil {
				return errors.Wrap(err, "failed to write config file")
			}
			return nil
		},
	}

	verifyCmd := &cobra.Command{
		Use:   "verify [module]",
		Short: "verify the config file is correct.",
		Long:  "verify the config file is correct based on the methods exported by modules.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				handler         handler
				targetFile      *os.File
				exportedConfigs []ExportedConfig
				confMap         map[string]any
				err             error
			)
			handler, err = getHandler(style)
			if err != nil {
				return err
			}
			if len(args) == 0 {
				exportedConfigs = m.exportedConfigs
			}
			if len(args) >= 1 {
				copy := make([]ExportedConfig, 0)
				for i := range m.exportedConfigs {
					for j := 0; j < len(args); j++ {
						if args[j] == m.exportedConfigs[i].Owner {
							copy = append(copy, m.exportedConfigs[i])
							break
						}
					}
				}
				exportedConfigs = copy
			}
			os.MkdirAll(filepath.Dir(targetFilePath), os.ModePerm)
			targetFile, err = os.OpenFile(targetFilePath,
				handler.flags(), os.ModePerm)
			if err != nil {
				return errors.Wrap(err, "failed to open config file")
			}
			defer targetFile.Close()
			bytes, err := ioutil.ReadAll(targetFile)
			if err != nil {
				return errors.Wrap(err, "failed to read config file")
			}
			err = handler.unmarshal(bytes, &confMap)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal config file")
			}
			for _, config := range exportedConfigs {
				if config.Validate == nil {
					continue
				}
				if err := config.Validate(confMap); err != nil {
					return errors.Wrap(err, "invalid config")
				}
			}
			return nil
		},
	}

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "manage configuration",
		Long:  "manage configuration, such as export a copy of default config.",
	}
	configCmd.PersistentFlags().StringVarP(
		&targetFilePath,
		"outputFile",
		"o",
		"./config/config.yaml",
		"The output file of exported config (alias of targetFile)",
	)
	configCmd.PersistentFlags().StringVarP(
		&targetFilePath,
		"targetFile",
		"t",
		"./config/config.yaml",
		"The targeted config file",
	)
	configCmd.PersistentFlags().StringVarP(
		&style,
		"style",
		"s",
		"yaml",
		"The output file style",
	)
	configCmd.AddCommand(initCmd)
	configCmd.AddCommand(verifyCmd)
	command.AddCommand(configCmd)
}

func loadValidators(k *KoanfAdapter, exportedConfigs []ExportedConfig) error {
	for _, config := range exportedConfigs {
		if config.Validate == nil {
			continue
		}
		k.validators = append(k.validators, config.Validate)
	}
	for _, f := range k.validators {
		if err := f(k.K.Raw()); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}
	}
	return nil
}

func getHandler(style string) (handler, error) {
	switch style {
	case "json":
		return rewriteHandler{codec: json.NewCodec(json.WithIndent("  "))}, nil
	case "yaml":
		return appendHandler{codec: yaml.Codec{}}, nil
	default:
		return nil, fmt.Errorf("unsupported config style %s", style)
	}
}

type handler interface {
	flags() int
	unmarshal(bytes []byte, o any) error
	write(file *os.File, configs []ExportedConfig, confMap map[string]any) error
}

type appendHandler struct {
	codec contract.Codec
}

func (y appendHandler) flags() int {
	return os.O_APPEND | os.O_CREATE | os.O_RDWR
}

func (y appendHandler) unmarshal(bytes []byte, o any) error {
	return y.codec.Unmarshal(bytes, o)
}

func (y appendHandler) write(file *os.File, configs []ExportedConfig, confMap map[string]any) error {
out:
	for i, config := range configs {
		for k := range config.Data {
			if _, ok := confMap[k]; ok {
				continue out
			}
		}
		if i != 0 {
			fmt.Fprintln(file, "")
		}
		bytes, err := y.codec.Marshal(config.Data)
		if err != nil {
			return err
		}
		if config.Comment != "" {
			_, err = fmt.Fprintln(file, "# "+config.Comment)
			if err != nil {
				return err
			}
		}
		_, err = file.Write(bytes)
		if err != nil {
			return err
		}
	}
	return nil
}

type rewriteHandler struct {
	codec contract.Codec
}

func (r rewriteHandler) flags() int {
	return os.O_CREATE | os.O_RDWR
}

func (r rewriteHandler) unmarshal(bytes []byte, o any) error {
	if len(bytes) == 0 {
		bytes = []byte("{}")
	}
	return r.codec.Unmarshal(bytes, o)
}

func (r rewriteHandler) write(file *os.File, configs []ExportedConfig, confMap map[string]any) error {
	if confMap == nil {
		confMap = make(map[string]any)
	}
	for _, exportedConfig := range configs {
		for k := range exportedConfig.Data {
			if _, ok := confMap[k]; ok {
				continue
			}
			confMap[k] = exportedConfig.Data[k]
		}
	}
	file.Seek(0, 0)
	data, err := r.codec.Marshal(confMap)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	fmt.Fprintln(file)
	return err
}
