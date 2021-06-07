package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DoNewsCode/core/di"
	"gopkg.in/yaml.v3"

	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/DoNewsCode/core/contract"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Module is the configuration module that bundles the reload watcher and exportConfig commands.
type Module struct {
	conf            *KoanfAdapter
	exportedConfigs []ExportedConfig
	dispatcher      contract.Dispatcher
}

// ConfigIn is the injection parameter for config.New.
type ConfigIn struct {
	di.In

	Conf            contract.ConfigAccessor
	Dispatcher      contract.Dispatcher `optional:"true"`
	ExportedConfigs []ExportedConfig    `group:"config"`
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
		m.conf.BindDispatcher(m.dispatcher)
		return m.conf.Watch(ctx)
	}, func(err error) {
		cancel()
	})
}

// ProvideCommand provides the config related command.
func (m Module) ProvideCommand(command *cobra.Command) {
	var (
		outputFile string
		style      string
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
				confMap         map[string]interface{}
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
				var copy = make([]ExportedConfig, 0)
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
			os.MkdirAll(filepath.Dir(outputFile), os.ModePerm)
			targetFile, err = os.OpenFile(outputFile,
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
	initCmd.Flags().StringVarP(
		&outputFile,
		"outputFile",
		"o",
		"./config/config.yaml",
		"The output file of exported config",
	)
	initCmd.Flags().StringVarP(
		&style,
		"style",
		"s",
		"yaml",
		"The output file style",
	)
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "manage configuration",
		Long:  "manage configuration, such as export a copy of default config.",
	}
	configCmd.AddCommand(initCmd)
	command.AddCommand(configCmd)
}

func getHandler(style string) (handler, error) {
	switch style {
	case "json":
		return jsonHandler{}, nil
	case "yaml":
		return yamlHandler{}, nil
	default:
		return nil, fmt.Errorf("unsupported config style %s", style)
	}
}

type yamlHandler struct {
}

func (y yamlHandler) flags() int {
	return os.O_APPEND | os.O_CREATE | os.O_RDWR
}

func (y yamlHandler) unmarshal(bytes []byte, o interface{}) error {
	return yaml.Unmarshal(bytes, o)
}

func (y yamlHandler) write(file *os.File, configs []ExportedConfig, confMap map[string]interface{}) error {
out:
	for _, config := range configs {
		for k := range config.Data {
			if _, ok := confMap[k]; ok {
				continue out
			}
		}
		bytes, err := yaml.Marshal(config.Data)
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
		_, err = fmt.Fprintln(file)
		if err != nil {
			return err
		}
	}
	return nil
}

type handler interface {
	flags() int
	unmarshal(bytes []byte, o interface{}) error
	write(file *os.File, configs []ExportedConfig, confMap map[string]interface{}) error
}

type jsonHandler struct {
}

func (y jsonHandler) flags() int {
	return os.O_CREATE | os.O_RDWR
}

func (y jsonHandler) unmarshal(bytes []byte, o interface{}) error {
	if len(bytes) == 0 {
		return nil
	}
	return json.Unmarshal(bytes, o)
}

func (y jsonHandler) write(file *os.File, configs []ExportedConfig, confMap map[string]interface{}) error {
	if confMap == nil {
		confMap = make(map[string]interface{})
	}
	for _, exportedConfig := range configs {
		if _, ok := confMap[exportedConfig.Owner]; ok {
			continue
		}
		for k := range exportedConfig.Data {
			confMap[k] = exportedConfig.Data[k]
		}
	}
	file.Seek(0, 0)
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(confMap)
	return err
}
