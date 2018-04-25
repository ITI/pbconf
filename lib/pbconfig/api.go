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

type CfgWebAPI struct {
	UseHSTS           bool
	Listen            string
	RequireClientCert bool
	ServerCert        string
	ServerKey         string
	TrustedCerts      string
	ClientCert        string `cfg_key:"optional"`
	ClientKey         string `cfg_key:"optional"`
	CipherSuites      string `cfg_key:"optional"`
	LogLevel          string `gcfg:"loglevel" cfg_key:"optional"`
}

//If program asks to "GetStringVar(x)" or GetBoolVar(x) where x does not exist in struct above, panic
func (c *CfgWebAPI) getvar(name string) interface{} {
	rv := reflect.TypeOf(*c)

	for i := 0; i < reflect.ValueOf(*c).NumField(); i++ {
		if rv.Field(i).Name == name {
			return reflect.ValueOf(*c).Field(i).Interface()
		}
	}
	panic("Field " + name + " not found in config struct.")
}

func (c *CfgWebAPI) CheckCfgFieldsExist() error {
	rt := reflect.TypeOf(*c)
	rv := reflect.ValueOf(*c)
	for i := 0; i < reflect.ValueOf(*c).NumField(); i++ {
		if rt.Field(i).Tag.Get("cfg_key") != "optional" && rt.Field(i).Type == reflect.TypeOf("string") && rv.Field(i).Interface() == "" {
			return errors.New("In section WebAPI, required config key " + rt.Field(i).Name + " not found")
		}
	}
	return nil
}

func (c *CfgWebAPI) GetStringVar(name string) string {
	s := c.getvar(name)
	return reflect.ValueOf(s).String()
}

func (c *CfgWebAPI) GetBoolVar(name string) bool {
	s := c.getvar(name)
	return reflect.ValueOf(s).Bool()
}
