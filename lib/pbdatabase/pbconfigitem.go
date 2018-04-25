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

type PbNodeConfigItem struct {
	NodeId int64 `json:"-"` //ignore this field in the json Marshaling
	ConfigItem
}

type ConfigItem struct {
	Key   string `validate:"nonzero"`
	Value string
}

func (configItem *PbNodeConfigItem) Exists(db AppDatabase) (bool, error) {
	return db.RowExists("SELECT EXISTS(SELECT 1 FROM ConfigItems JOIN ConfigLines ON ConfigItems.id = ConfigLines.configItem AND ConfigItems.key=? AND node=? LIMIT 1)", "ConfigItemExists", configItem.Key, configItem.NodeId)
}

func (configItem *PbNodeConfigItem) Create(db AppDatabase) error {
	//verify node existence
	tmp_node := PbNode{Id: configItem.NodeId}
	exists, err := tmp_node.ExistsById(db)
	if err != nil {
		db.log.Debug("1. Create Node ConfigElement Error: %s", err)
		return err
	}
	if !exists {
		db.log.Debug("CreateNodeConfigElement Node existence error")
		return errors.New("Could not find the node to add config item to.")
	}

	transaction, err := db.Begin()
	if err != nil {
		db.log.Debug("2. Create Node ConfigElement Error: %s", err.Error())
		return err
	}
	err = configItem.createtransaction(db, transaction)
	if err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit()
}

func (configItem *PbNodeConfigItem) Get(db AppDatabase) error {
	err := db.QueryRow("SELECT value FROM ConfigItems JOIN ConfigLines ON ConfigItems.id = ConfigLines.configItem AND ConfigItems.key=? AND node=?",
		configItem.Key, configItem.NodeId).Scan(&configItem.Value)
	if err != nil {
		db.log.Debug("GetNodeConfigElement Error: %s", err.Error())
		return err
	}
	return nil
}

func (configItem *PbNodeConfigItem) CreateOrUpdate(db AppDatabase) error {
	transaction, err := db.Begin()
	if err != nil {
		db.log.Debug("CreateOrUpdate Node config Item Error: %s", err.Error())
		return err
	}
	err = configItem.createorupdatetransaction(db, transaction)
	if err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit()
}

func (configItem *PbNodeConfigItem) createorupdatetransaction(db AppDatabase, transaction *sql.Tx) error {
	exists, err := configItem.Exists(db)
	if err != nil {
		db.log.Debug("CreateOrUpdateNodeConfigElement Error: %s", err.Error())
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

func (configItem *PbNodeConfigItem) createtransaction(db AppDatabase, transaction *sql.Tx) error {
	if err := validator.Validate(configItem); err != nil {
		db.log.Debug("1. createtransaction NodeConfigElement Validation Error: %s", err.Error())
		return err
	}
	res, err := transaction.Exec("INSERT INTO ConfigItems VALUES(?, ?, ?)", nil, configItem.Key, configItem.Value)
	if err != nil {
		db.log.Debug("2. createtransaction NodeConfigElement Error: %s", err.Error())
		return err
	}
	confId, err := res.LastInsertId()
	if err != nil {
		db.log.Debug("3. createtransaction NodeConfigElement Error: %s", err.Error())
		return err
	}
	_, err = transaction.Exec("INSERT INTO ConfigLines VALUES(?, ?)", configItem.NodeId, confId)
	if err != nil {
		db.log.Debug("4. createtransaction NodeConfigElement Error: %s", err.Error())
		return err
	}
	return nil
}
func (configItem *PbNodeConfigItem) updatetransaction(db AppDatabase, transaction *sql.Tx) error {
	//find the config item id for the given key-value pair
	var configId int64
	err := db.QueryRow("SELECT id FROM ConfigItems JOIN ConfigLines ON ConfigItems.id = ConfigLines.configItem AND ConfigItems.key=? AND node=?",
		configItem.Key, configItem.NodeId).Scan(&configId)
	if err != nil {
		db.log.Debug("1. updatetransaction NodeConfigElement Error: %s", err.Error())
		return err
	}
	_, err = transaction.Exec("REPLACE INTO ConfigItems VALUES(?, ?, ?)", configId, configItem.Key, configItem.Value)
	if err != nil {
		db.log.Debug("2. updatetransaction NodeConfigElement Error: %s", err.Error())
	}
	return err
}

func (configItem *PbNodeConfigItem) Delete(db AppDatabase) error {
	transaction, err := db.Begin()
	if err != nil {
		db.log.Debug("DeleteNodeConfigElement Error: %s", err.Error())
		return err
	}
	var configId int64
	err = transaction.QueryRow(`SELECT id FROM ConfigItems WHERE Key IN
		(SELECT ConfigItems.Key FROM ConfigItems JOIN ConfigLines ON ConfigItems.id = ConfigLines.configItem AND node=? AND ConfigItems.Key=?)`,
		configItem.NodeId, configItem.Key).Scan(&configId)
	if err != nil {
		transaction.Rollback()
		db.log.Debug("1. DeleteNodeConfigElement Error: %s", err.Error())
		return err
	}

	_, err = transaction.Exec(`DELETE FROM ConfigItems WHERE id=?`, configId)
	if err != nil {
		transaction.Rollback()
		db.log.Debug("2. DeleteNodeConfigElement Error: %s", err.Error())
		return err
	}

	_, err = transaction.Exec("DELETE FROM ConfigLines WHERE node=? AND configItem=?", configItem.NodeId, configId)
	if err != nil {
		transaction.Rollback()
		db.log.Debug("3. DeleteNodeConfigElement Error: %s", err.Error())
		return err
	}
	transaction.Commit()
	return nil
}
