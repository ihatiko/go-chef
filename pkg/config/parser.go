package config

import (
	"github.com/spf13/viper"
	"reflect"
	"strings"
)

func (e *Config) Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	if err := e.Viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			// 	do nothing
		default:
			return err
		}
	}
	_ = e.Viper.Unmarshal(rawVal, opts...)
	e.readEnvs(rawVal)
	return e.Viper.Unmarshal(rawVal, opts...)
}

func (e *Config) readEnvs(rawVal interface{}) {
	e.Viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	e.bindEnvs(rawVal)
}

func (e *Config) bindEnvs(in interface{}, prev ...string) {
	ifv := reflect.ValueOf(in)
	if ifv.Kind() == reflect.Ptr {
		ifv = ifv.Elem()
	}
	for i := 0; i < ifv.NumField(); i++ {
		fv := ifv.Field(i)
		if fv.Kind() == reflect.Ptr {
			if fv.IsZero() {
				fv = reflect.New(fv.Type().Elem()).Elem()
			} else {
				fv = fv.Elem()
			}
		}
		t := ifv.Type().Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if ok {
			if tv == ",squash" {
				e.bindEnvs(fv.Interface(), prev...)
				continue
			}
		} else {
			tv = t.Name
		}
		switch fv.Kind() {
		case reflect.Struct:
			e.bindEnvs(fv.Interface(), append(prev, tv)...)
		case reflect.Map:
			iter := fv.MapRange()
			for iter.Next() {
				if key, ok := iter.Key().Interface().(string); ok {
					e.bindEnvs(iter.Value().Interface(), append(prev, tv, key)...)
				}
			}
		default:
			env := strings.Join(append(prev, tv), ".")
			_ = e.Viper.BindEnv(env)
		}
	}
}
