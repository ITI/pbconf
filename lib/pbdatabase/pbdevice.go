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
	"time"
)

type PbDevice struct {
	Id          int64
	Name        string `validate:"nonzero"`
	ParentNode  *int64
	ConfigItems []ConfigItem
}

func (dev *PbDevice) ExistsByName(db AppDatabase) (bool, error) {
	return db.RowExists("SELECT EXISTS(SELECT 1 FROM Devices WHERE name = ? LIMIT 1)", "DeviceExistByName", dev.Name)
}

func (dev *PbDevice) ExistsById(db AppDatabase) (bool, error) {
	return db.RowExists("SELECT EXISTS(SELECT 1 FROM Devices WHERE id = ? LIMIT 1)", "DeviceExistById", dev.Id)
}

func (dev *PbDevice) Create(db AppDatabase) error {
	transaction, err := db.Begin()
	if err != nil {
		db.log.Debug("Create Device Transaction begin Error: %s", err.Error())
		return err
	}
	err = dev.createDeviceTableTransaction(db, transaction)
	if err != nil {
		transaction.Rollback()
		return err
	}
	if err = dev.createConfigItemsTransaction(db, transaction); err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit()
}

func (dev *PbDevice) Update(db AppDatabase) error {
	exists, err := dev.ExistsById(db)
	if err != nil {
		db.log.Debug("Error while checking existence of device in Devices table")
		return err
	}
	if !exists {
		db.log.Debug("Device does not exist in Device table")
		return errors.New("Device does not exist")
	}
	transaction, err := db.Begin()
	if err != nil {
		db.log.Debug("Update Device Transaction Error: %s", err.Error())
		return err
	}
	if err = dev.updateDeviceTableTransaction(db, transaction); err != nil {
		transaction.Rollback()
		return err
	}
	if err = dev.createOrUpdateConfigItemsTransaction(db, transaction); err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit()
}

func (dev *PbDevice) Delete(db AppDatabase) error {
	transaction, err := db.Begin()
	if err != nil {
		db.log.Debug("Delete Device Error: %s", err.Error())
		return err
	}
	//delete dependencies first
	if err = dev.deleteConfigItemsTransaction(db, transaction); err != nil {
		transaction.Rollback()
		return err
	}
	//delete device table row
	if err = dev.deleteDeviceTableTransaction(db, transaction); err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit()
}

func (dev *PbDevice) Get(db AppDatabase) error {
	//var device pbdata.PbDevice
	err := db.QueryRow("SELECT name, parentNode FROM Devices WHERE id=?", dev.Id).Scan(&dev.Name, &dev.ParentNode)
	if err != nil {
		db.log.Debug("GetDevice Error: %s", err.Error())
		return err
	}
	return dev.getConfigItems(db)
}

func (dev *PbDevice) GetByName(db AppDatabase) error {
	err := db.QueryRow("SELECT id, parentNode FROM devices WHERE name=?", dev.Name).Scan(&dev.Id, &dev.ParentNode)
	if err != nil {
		db.log.Debug("GetByName Device Error: %s", err.Error())
		return err
	}
	return dev.getConfigItems(db)
}

func (dev *PbDevice) GetConnectionString(db AppDatabase, subsys string) (string, string, error) {

	tk := subsys + "transport"
	lk := subsys + "location"

	transItem := PbDeviceConfigItem{DeviceId: dev.Id, ConfigItem: ConfigItem{Key: tk}}
	locItem := PbDeviceConfigItem{DeviceId: dev.Id, ConfigItem: ConfigItem{Key: lk}}

	tr_exists, err := transItem.Exists(db)
	if err != nil {
		return "", "", err
	}
	loc_exists, err := locItem.Exists(db)
	if err != nil {
		return "", "", err
	}
	if !tr_exists || !loc_exists {
		return "", "", errors.New("No transport or location config item found")
	}

	if err := transItem.Get(db); err != nil {
		return "", "", err
	}

	if err := locItem.Get(db); err != nil {
		return "", "", err
	}
	return transItem.Value, locItem.Value, nil
}

/*************** Device, DeviceConfigItems, DeviceConfigLines table access helper functions ********************/
func (dev *PbDevice) createDeviceTableTransaction(db AppDatabase, transaction *sql.Tx) error {
	if dev.ParentNode == nil {
		db.log.Debug("Cannot create a device without a parent node")
		return errors.New("CreateDevice Error: No parent node specified")
	}

	res, err := transaction.Exec("INSERT INTO Devices VALUES(?, ?, ?)", nil, dev.Name, *dev.ParentNode)
	if err != nil {
		db.log.Debug("1. CreateDevice Error: %s", err.Error())
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		db.log.Debug("2. CreateDevice Error: %s", err.Error())
		return err
	}
	dev.Id = id
	return nil
}

//createConfigItemsTransaction does NOT check the existence of the device in the database. It may be part of
// the transaction to create the device.
func (dev *PbDevice) createConfigItemsTransaction(db AppDatabase, transaction *sql.Tx) error {
	for _, confItem := range dev.ConfigItems {
		pbConfItem := PbDeviceConfigItem{dev.Id, confItem}
		if err := pbConfItem.createtransaction(db, transaction); err != nil {
			return err
		}
	}
	return nil
}

func (dev *PbDevice) deleteConfigItemsTransaction(db AppDatabase, transaction *sql.Tx) error {
	_, err := transaction.Exec("DELETE FROM DeviceConfigItems WHERE id IN (SELECT DeviceConfigItems.id FROM DeviceConfigItems JOIN DeviceConfigLines ON DeviceConfigItems.id = DeviceConfigLines.deviceconfigItem AND device=?)", dev.Id)
	if err != nil {
		db.log.Debug("delete from ConfigItems table Error: %s", err.Error())
		return err
	}
	_, err = transaction.Exec("DELETE FROM DeviceConfigLines WHERE device=?", dev.Id)
	if err != nil {
		db.log.Debug("DeleteDeviceConfig DeviceConfigLines Error: %s", err.Error())
		return err
	}
	return nil
}

func (dev *PbDevice) deleteDeviceTableTransaction(db AppDatabase, transaction *sql.Tx) error {
	_, err := transaction.Exec("DELETE FROM Devices WHERE id=?", dev.Id)
	if err != nil {
		db.log.Debug("deleteDeviceTableTransaction Error: %s", err.Error())
	}
	return err
}

func (dev *PbDevice) getConfigItems(db AppDatabase) error {
	rows, err := db.Query("SELECT key, value FROM DeviceConfigItems JOIN DeviceConfigLines ON DeviceConfigItems.id = DeviceConfigLines.deviceconfigItem AND device=?", dev.Id)
	if err != nil {
		db.log.Debug("Select query, getConfigItems Error: %s", err.Error())
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var confItem ConfigItem
		err = rows.Scan(&confItem.Key, &confItem.Value)
		if err != nil {
			db.log.Debug("row scan, getConfigItems Error: %s", err.Error())
			return err
		}
		dev.ConfigItems = append(dev.ConfigItems, confItem)
	}
	if err := rows.Err(); err != nil {
		db.log.Debug("getConfigItems Error: %s", err.Error())
		return err
	}
	return nil
}

func (dev *PbDevice) updateDeviceTableTransaction(db AppDatabase, transaction *sql.Tx) error {
	if dev.ParentNode == nil {
		db.log.Debug("Cannot update a device without a parent node")
		return errors.New("updateDeviceTableTransaction Error: No parent node specified")
	}

	_, err := transaction.Exec("INSERT OR REPLACE INTO Devices VALUES(?, ?, ?)", dev.Id, dev.Name, *dev.ParentNode)
	if err != nil {
		db.log.Debug("updateDeviceTableTransaction Error: %s", err.Error())
	}
	return err
}

func (dev *PbDevice) createOrUpdateConfigItemsTransaction(db AppDatabase, transaction *sql.Tx) error {
	for _, confItem := range dev.ConfigItems {
		dbConfItem := PbDeviceConfigItem{dev.Id, confItem}
		if err := dbConfItem.createorupdatetransaction(db, transaction); err != nil {
			return err
		}
	}
	return nil
}

/******************************* Device Table only access functions ********************************************************/
func (db AppDatabase) GetDevices() ([]PbDevice, error) {
	var deviceList []PbDevice
	rows, err := db.Query("SELECT id, name, parentNode FROM devices")
	if err != nil {
		db.log.Debug("1.GetDevices Error: %s", err.Error())
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var device PbDevice
		if err := rows.Scan(&device.Id, &device.Name, &device.ParentNode); err != nil {
			db.log.Debug("2.GetDevices Error: %s", err.Error())
			return nil, err
		}
		deviceList = append(deviceList, device)
	}
	if err := rows.Err(); err != nil {
		db.log.Debug("3.GetDevices Error: %s", err.Error())
		return nil, err
	}
	return deviceList, nil
}

func (db AppDatabase) GetNodeDevices(parentNodeId int64) ([]PbDevice, error) {
	var deviceList []PbDevice
	rows, err := db.Query("SELECT id, name, parentNode FROM devices WHERE parentNode=?", parentNodeId)
	if err != nil {
		db.log.Debug("1.GetNodeDevices Error: %s", err.Error())
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var device PbDevice
		if err := rows.Scan(&device.Id, &device.Name, &device.ParentNode); err != nil {
			db.log.Debug("2.GetNodeDevices Error: %s", err.Error())
			return nil, err
		}
		deviceList = append(deviceList, device)
	}
	if err := rows.Err(); err != nil {
		db.log.Debug("3.GetNodeDevices Error: %s", err.Error())
		return nil, err
	}
	return deviceList, nil
}

func (db AppDatabase) GetDevicesLastModified() (string, error) {
	exists, err := db.RowExists("SELECT EXISTS(SELECT 1 FROM Log WHERE tableName='Devices' LIMIT 1)", "GetDevicesLastModified")
	if err != nil {
		db.log.Debug("GetDevicesLastModified: Got error while seeing if row exists, Error: %s", err.Error())
		return time.RFC1123, err
	}
	if !exists {
		return time.RFC1123, nil //return some constant timestamp if no devices
	}
	var logdatetime time.Time
	err = db.QueryRow("SELECT lastModified FROM Log WHERE tableName='Devices'").Scan(&logdatetime)
	if err != nil {
		db.log.Debug("GetDevicesLastModified: Got error from querying database row Error: %s", err.Error())
	}
	return logdatetime.Format(time.RFC1123), err
}
