package pbconfig

/***********************************************************************
   Copyright 2018 Information Trust Institute

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
***********************************************************************/

import (
	"errors"
	"reflect"
)

type CfgTranslator struct {
	SocketDir    string   `gcfg:"socketdir"`
	LogLevel     string   `gcfg:"loglevel" cfg_key:"optional"`
	ModuleDir    string   `gcfg:"moduledir"`
	TransModules []string `gcfg:"module" cfg_key:"optional"`
	ForceVerify  bool     `gcfg:"forceverify" cfg_key:"optional"`
}

func (c *CfgTranslator) String() string {
	return "{}"
}

func (c *CfgTranslator) CheckCfgFieldsExist() error {
	rt := reflect.TypeOf(*c)
	rv := reflect.ValueOf(*c)
	for i := 0; i < reflect.ValueOf(*c).NumField(); i++ {
		if rt.Field(i).Tag.Get("cfg_key") != "optional" && rt.Field(i).Type == reflect.TypeOf("string") && rv.Field(i).Interface() == "" {
			return errors.New("In section Translator, required config key " + rt.Field(i).Name + " not found")
		}
	}
	return nil
}

type CfgDriverOpts struct {
	Hash     string `gcfg:"hash" cfg_key:"optional"`
	HashType string `gcfg:"type" cfg_key:"optional"`
	Path     string `gcfg:"path" cfg_key:"optional"`
}

func (c *CfgDriverOpts) CheckCfgFieldsExist() error {
	rt := reflect.TypeOf(*c)
	rv := reflect.ValueOf(*c)
	for i := 0; i < reflect.ValueOf(*c).NumField(); i++ {
		if rt.Field(i).Tag.Get("cfg_key") != "optional" && rt.Field(i).Type == reflect.TypeOf("string") && rv.Field(i).Interface() == "" {
			return errors.New("required config key " + rt.Field(i).Name + " not found")
		}
	}
	return nil
}
