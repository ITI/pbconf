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

type CfgConBroker struct {
	PubKey   string
	PrivKey  string
	Listen   string
	LogLevel string `gcfg:"loglevel" cfg_key:"optional"`
}

func (c *CfgConBroker) CheckCfgFieldsExist() error {
	rt := reflect.TypeOf(*c)
	rv := reflect.ValueOf(*c)
	for i := 0; i < reflect.ValueOf(*c).NumField(); i++ {
		if rt.Field(i).Tag.Get("cfg_key") != "optional" && rt.Field(i).Type == reflect.TypeOf("string") && rv.Field(i).Interface() == "" {
			return errors.New("In section Broker, required config key " + rt.Field(i).Name + " not found")
		}
	}
	return nil
}

func (c *CfgConBroker) String() string {
	return "{}"
}
