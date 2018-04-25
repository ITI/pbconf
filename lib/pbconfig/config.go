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
	"reflect"

	gcfg "gopkg.in/gcfg.v1"
)

type Cfg interface {
	CheckCfgFieldsExist() error
}

type Config struct {
	Global      cfgGlobal                 `gcfg:"global"`
	SvcManager  cfgServiceManager         `gcfg:"service-manager"`
	WebAPI      CfgWebAPI                 `gcfg:"webapi"`
	WebUI       CfgWebUI                  `gcfg:"web-ui"`
	Policy      CfgPolicy                 `gcfg:"policy"`
	ChMgmt      cfgChange                 `gcfg:"change"`
	Broker      CfgConBroker              `gcfg:"broker"`
	Translation CfgTranslator             `gcfg:"translation"`
	DriverOpts  map[string]*CfgDriverOpts `gcfg:"driveroptions"`

	path string
}

func (c *Config) LoadConfig(cfgfile string) error {
	c.path = cfgfile
	return gcfg.ReadFileInto(c, cfgfile)
}

func (c *Config) CheckCfgFieldsExist() error {
	rt := reflect.TypeOf(*c)
	rv := reflect.ValueOf(*c)
	for i := 0; i < reflect.ValueOf(*c).NumField(); i++ {
		if rt.Field(i).PkgPath != "" { //if struct field is unexported, continue
			continue
		}
		_, ok := reflect.PtrTo(rt.Field(i).Type).MethodByName("CheckCfgFieldsExist") //check if struct field we are looking at has the function defined
		if ok {
			ptrCfgItem := reflect.New(rt.Field(i).Type)                                       //create new pointer to type cfgGlobal/cfgWebAPI/etc
			valueCfgItem := ptrCfgItem.Elem()                                                 //create variable tp value of the pointer
			valueCfgItem.Set(rv.Field(i))                                                     // set value of variable to our value
			retVals := ptrCfgItem.MethodByName("CheckCfgFieldsExist").Call([]reflect.Value{}) //call CheckCfgFieldsExist function on the pointer receiver
			err, found := retVals[0].Interface().(error)
			if found {
				return err
			}
		}

	}
	return nil
}

func (c *Config) Path() string {
	return c.path
}

func NewConfig(cfgfile string) (*Config, error) {
	c := new(Config)
	if err := c.LoadConfig(cfgfile); err != nil {
		return c, err
	}
	err := c.CheckCfgFieldsExist()
	return c, err
}
