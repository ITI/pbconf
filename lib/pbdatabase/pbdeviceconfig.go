package pbdatabase

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
	"database/sql"
	"errors"

	validator "gopkg.in/validator.v2"
)

type PbDeviceConfigItem struct {
	DeviceId int64 `json:"-"` //ignore this field in the json Marshaling
	ConfigItem
}

func (configItem *PbDeviceConfigItem) Exists(db AppDatabase) (bool, error) {
	return db.RowExists("SELECT EXISTS(SELECT 1 FROM DeviceConfigItems JOIN DeviceConfigLines ON DeviceConfigItems.id = DeviceConfigLines.deviceconfigItem AND DeviceConfigItems.key=? AND device=? LIMIT 1)", "ConfigItemExists", configItem.Key, configItem.DeviceId)
}

func (configItem *PbDeviceConfigItem) Create(db AppDatabase) error {
	transaction, err := db.Begin()
	if err != nil {
		db.log.Debug("CreateDeviceConfigElement Error: %s", err.Error())
		return err
	}

	err = configItem.confirmdevcreatetransaction(db, transaction)
	if err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit()
}

func (configItem *PbDeviceConfigItem) Get(db AppDatabase) error {
	err := db.QueryRow("SELECT value FROM DeviceConfigItems JOIN DeviceConfigLines ON DeviceConfigItems.id = DeviceConfigLines.deviceConfigItem AND DeviceConfigItems.key=? AND device=?",
		configItem.Key, configItem.DeviceId).Scan(&configItem.Value)
	if err != nil {
		db.log.Debug("GetDeviceConfigElement Error: %s", err.Error())
	}
	return err
}

func (configItem *PbDeviceConfigItem) Update(db AppDatabase) error {
	transaction, err := db.Begin()
	if err != nil {
		db.log.Debug("UpdateDeviceConfigElement Error: %s", err.Error())
		return err
	}
	err = configItem.updatetransaction(db, transaction)
	if err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit()
}

func (configItem *PbDeviceConfigItem) Delete(db AppDatabase) error {
	transaction, err := db.Begin()
	if err != nil {
		db.log.Debug("DeleteDeviceConfigElement Error: %s", err.Error())
		return err
	}
	err = configItem.deletetransaction(db, transaction)
	if err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit()
}

func (configItem *PbDeviceConfigItem) deletetransaction(db AppDatabase, transaction *sql.Tx) error {
	var configId int64
	err := transaction.QueryRow(`SELECT id FROM DeviceConfigItems WHERE Key IN
		(SELECT DeviceConfigItems.Key FROM DeviceConfigItems JOIN DeviceConfigLines ON DeviceConfigItems.id = DeviceConfigLines.deviceConfigItem AND device=? AND DeviceConfigItems.Key=?)`,
		configItem.DeviceId, configItem.Key).Scan(&configId)
	if err != nil {
		db.log.Debug("1. DeleteDeviceConfigElement Error: %s", err.Error())
		return err
	}

	_, err = transaction.Exec(`DELETE FROM DeviceConfigItems WHERE id=?`, configId)
	if err != nil {
		db.log.Debug("2. DeleteDeviceConfigElement Error: %s", err.Error())
		return err
	}

	_, err = transaction.Exec("DELETE FROM DeviceConfigLines WHERE device=? AND deviceconfigItem=?", configItem.DeviceId, configId)
	if err != nil {
		db.log.Debug("3. DeleteDeviceConfigElement Error: %s", err.Error())
		return err
	}
	return nil
}

func (configItem *PbDeviceConfigItem) createorupdatetransaction(db AppDatabase, transaction *sql.Tx) error {
	exists, err := configItem.Exists(db)
	if err != nil {
		db.log.Debug("CreateOrUpdateDeviceConfigElement Error: %s", err.Error())
		return err
	}

	if !exists {
		err = configItem.createtransaction(db, transaction)
		if err != nil {
			return err
		}
	} else {
		err = configItem.updatetransaction(db, transaction)
		if err != nil {
			return err
		}
	}
	return nil
}

func (configItem *PbDeviceConfigItem) confirmdevcreatetransaction(db AppDatabase, transaction *sql.Tx) error {
	//verify Device existence
	tmp_device := PbDevice{Id: configItem.DeviceId}
	exists, err := tmp_device.ExistsById(db)
	if err != nil {
		db.log.Debug("1. Confirm Device Configitem Error: %s", err)
		return err
	}
	if !exists {
		db.log.Debug("In Confirm Device Configitem, Device with id %v does not exist", tmp_device.Id)
		return errors.New("Device does not exist")
	}

	err = configItem.createtransaction(db, transaction)
	return err
}

func (configItem *PbDeviceConfigItem) createtransaction(db AppDatabase, transaction *sql.Tx) error {
	if err := validator.Validate(configItem); err != nil {
		db.log.Debug("1. Create DeviceConfigItem Validation error: %s", err.Error())
		return err
	}
	res, err := transaction.Exec("INSERT INTO DeviceConfigItems VALUES(?, ?, ?)", nil, configItem.Key, configItem.Value)
	if err != nil {
		db.log.Debug("2. Create Device Configitem Error: %s", err.Error())
		return err
	}
	confId, err := res.LastInsertId()
	if err != nil {
		db.log.Debug("3. Create Device Configitem Error: %s", err.Error())
		return err
	}
	_, err = transaction.Exec("INSERT INTO DeviceConfigLines VALUES(?, ?)", configItem.DeviceId, confId)
	if err != nil {
		db.log.Debug("4. Create Device Configitem Error: %s", err.Error())
		return err
	}
	return nil
}

func (configItem *PbDeviceConfigItem) updatetransaction(db AppDatabase, transaction *sql.Tx) error {
	var dbConfItem PbDeviceConfigItem
	var configId int64
	err := db.QueryRow("SELECT id, key, value FROM DeviceConfigItems JOIN DeviceConfigLines ON DeviceConfigItems.id = DeviceConfigLines.deviceConfigItem AND DeviceConfigItems.key=? AND device=?",
		configItem.Key, configItem.DeviceId).Scan(&configId, &dbConfItem.Key, &dbConfItem.Value)
	if err != nil {
		db.log.Debug("UpdateDeviceConfigElement Error: %s", err.Error())
		return err
	}
	_, err = transaction.Exec("REPLACE INTO DeviceConfigItems VALUES(?, ?, ?)", configId, dbConfItem.Key, configItem.Value)
	if err != nil {
		db.log.Debug("UpdateDeviceConfigElement Error: %s", err.Error())
		return err
	}
	return nil
}
