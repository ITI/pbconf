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
	"strconv"
	"time"
)

type PbNode struct {
	Id          int64
	Name        string `validate:"nonzero"`
	ConfigItems []ConfigItem
}

func (node *PbNode) ExistsByName(db AppDatabase) (bool, error) {
	return db.RowExists("SELECT EXISTS(SELECT 1 FROM Nodes WHERE name = ? LIMIT 1)", "NodeExistByName", node.Name)
}

func (node *PbNode) ExistsById(db AppDatabase) (bool, error) {
	return db.RowExists("SELECT EXISTS(SELECT 1 FROM Nodes WHERE id = ? LIMIT 1)", "NodeExistById", node.Id)
}

func (node *PbNode) Get(db AppDatabase) error {
	err := db.QueryRow("SELECT name FROM Nodes WHERE id=?", node.Id).Scan(&node.Name)
	if err != nil {
		db.log.Debug("GetNode Error: %s", err.Error())
		return err
	}
	return node.getConfigItems(db)
}

func (node *PbNode) GetByName(db AppDatabase) error {
	err := db.QueryRow("SELECT id FROM Nodes WHERE name=?", node.Name).Scan(&node.Id)
	if err != nil {
		db.log.Debug("GetByName Node error: %s", err.Error())
		return err
	}
	return node.getConfigItems(db)
}

func (node *PbNode) Create(db AppDatabase) error {
	transaction, err := db.Begin()
	if err != nil {
		db.log.Debug("Create Node Transaction Error: %s", err.Error())
		return err
	}
	if err = node.createNodeTableTransaction(db, transaction); err != nil {
		transaction.Rollback()
		return err
	}
	if err = node.createConfigItemsTransaction(db, transaction); err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit()
}

func (node *PbNode) Update(db AppDatabase) error {
	exists, err := node.ExistsById(db)
	if err != nil {
		db.log.Debug("Error while checking existence of node in node table")
		return err
	}
	if !exists {
		db.log.Debug("Node does not exist in node table")
		return errors.New("Node does not exist")
	}
	transaction, err := db.Begin()
	if err != nil {
		db.log.Debug("Update Node Transaction Error: %s", err.Error())
		return err
	}
	// disallow changing name/id of node
	if err = node.updateNodeTableTransaction(db, transaction); err != nil {
		transaction.Rollback()
		return err
	}

	if err = node.createOrUpdateConfigItemsTransaction(db, transaction); err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit()
}

func (node *PbNode) Delete(db AppDatabase) error {
	transaction, err := db.Begin()
	if err != nil {
		db.log.Debug("DeleteNode Error: %s", err.Error())
		return err
	}
	//delete dependencies first
	if err = node.deleteConfigItemsTransaction(db, transaction); err != nil {
		transaction.Rollback()
		return err
	}
	//delete node table row
	if err = node.deleteNodeTableTransaction(db, transaction); err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit()
}

/************************** Node, ConfigItems, ConfigLines Table accessor helper functions *****************************/
func (node *PbNode) getConfigItems(db AppDatabase) error {
	rows, err := db.Query("SELECT key, value FROM ConfigItems JOIN ConfigLines ON ConfigItems.id = ConfigLines.configItem AND node=?", node.Id)
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
		node.ConfigItems = append(node.ConfigItems, confItem)
	}
	if err := rows.Err(); err != nil {
		db.log.Debug("getConfigItems Error: %s", err.Error())
		return err
	}
	return nil
}

func (node *PbNode) createNodeTableTransaction(db AppDatabase, transaction *sql.Tx) error {
	res, err := transaction.Exec("INSERT INTO Nodes VALUES(?, ?)", nil, node.Name)
	if err != nil {
		db.log.Debug("createNodeTableTransaction INSERT Error: %s", err.Error())
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		db.log.Debug("createNodeTableTransaction LastInsertId Error: %s", err.Error())
		return err
	}
	node.Id = id
	return nil
}

func (node *PbNode) createConfigItemsTransaction(db AppDatabase, transaction *sql.Tx) error {
	for _, confItem := range node.ConfigItems {
		pbConfItem := PbNodeConfigItem{node.Id, confItem}
		if err := pbConfItem.createtransaction(db, transaction); err != nil {
			return err
		}
	}
	return nil
}

func (node *PbNode) updateNodeTableTransaction(db AppDatabase, transaction *sql.Tx) error {
	var dbName string
	err := db.QueryRow("SELECT name FROM Nodes WHERE id=?", node.Id).Scan(&dbName)
	if err != nil {
		return err
	}
	if dbName != node.Name {
		db.log.Debug("updateNodeTableTransaction Error: Cannot update node name")
		return errors.New("Cannot allow node name to be changed.")
	}
	return nil
}

func (node *PbNode) createOrUpdateConfigItemsTransaction(db AppDatabase, transaction *sql.Tx) error {
	for _, confItem := range node.ConfigItems {
		dbConfItem := PbNodeConfigItem{node.Id, confItem}
		if err := dbConfItem.createorupdatetransaction(db, transaction); err != nil {
			return err
		}
	}
	return nil
}

func (node *PbNode) deleteConfigItemsTransaction(db AppDatabase, transaction *sql.Tx) error {
	_, err := transaction.Exec("DELETE FROM ConfigItems WHERE id IN (SELECT ConfigItems.id FROM ConfigItems JOIN ConfigLines ON ConfigItems.id = ConfigLines.configItem AND node=?)", node.Id)
	if err != nil {
		db.log.Debug("delete from ConfigItems table Error: %s", err.Error())
		return err
	}
	_, err = transaction.Exec("DELETE FROM ConfigLines WHERE node=?", node.Id)
	if err != nil {
		db.log.Debug("delete from ConfigLines table Error: %s", err.Error())
		return err
	}
	return nil
}

func (node *PbNode) deleteNodeTableTransaction(db AppDatabase, transaction *sql.Tx) error {
	_, err := transaction.Exec("DELETE FROM Nodes WHERE id=?", node.Id)
	if err != nil {
		db.log.Debug("DeleteNode Error: %s", err.Error())
	}
	return err
}

/****************** Node Name, Id returned functions *****************************/
func (db AppDatabase) GetNodes() ([]PbNode, error) {
	var nodeList []PbNode
	rows, err := db.Query("SELECT id, name FROM Nodes")
	if err != nil {
		db.log.Debug("GetNodes Error: %s", err.Error())
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var node PbNode
		if err := rows.Scan(&node.Id, &node.Name); err != nil {
			db.log.Debug("GetNodes Error: %s", err.Error())
			return nil, err
		}
		nodeList = append(nodeList, node)
	}
	if err := rows.Err(); err != nil {
		db.log.Debug("GetNodes Error: %s", err.Error())
		return nil, err
	}
	return nodeList, nil
}

func (db AppDatabase) GetNodesLastModified() (string, error) {
	var logdatetime time.Time
	err := db.QueryRow("SELECT lastModified FROM Log WHERE tableName='Nodes'").Scan(&logdatetime)
	if err != nil {
		db.log.Debug("GetNodesLastModified: Got error from querying database row Error: %s", err.Error())
	}
	return logdatetime.Format(time.RFC1123), err
}

func (n *PbNode) GetChildNodes(db AppDatabase) ([]PbNode, error) {
	var nodeList []PbNode
	rows, err := db.Query("SELECT id, name FROM Nodes WHERE name <> ?", n.Name)
	if err != nil {
		db.log.Debug("GetNodes Error: %s", err.Error())
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var node PbNode
		if err := rows.Scan(&node.Id, &node.Name); err != nil {
			db.log.Debug("GetNodes Error: %s", err.Error())
			return nil, err
		}
		nodeList = append(nodeList, node)
	}
	if err := rows.Err(); err != nil {
		db.log.Debug("GetNodes Error: %s", err.Error())
		return nil, err
	}
	return nodeList, nil
}

/********************* Utility functions that return specific property value of the node **************/
// GetTimeoutMultiplier returns the multiplier to the reportTime && poll time at which time we assume the node
// connection to the node above or below has failed
func (db AppDatabase) GetTimeoutMultiplier(nodeId int64) int {
	multiplier := 3
	cfgItem := PbNodeConfigItem{NodeId: nodeId, ConfigItem: ConfigItem{Key: "HeartBeatTimeoutMultiplier"}}
	if err := cfgItem.Get(db); err != nil {
		db.log.Debug("Could not recover a timeout to check on child nodes, using default of 3")
	} else {
		multiplier, err = strconv.Atoi(cfgItem.Value)
		if err != nil {
			multiplier = 3
		}
	}
	return multiplier
}

func (db AppDatabase) GetDurationForTimer(nodeId int64, timerKey string) (time.Duration, error) {
	var duration time.Duration
	cfgItem := PbNodeConfigItem{NodeId: nodeId, ConfigItem: ConfigItem{Key: timerKey}}
	if err := cfgItem.Get(db); err != nil {
		return duration, err
	}

	db.log.Info("Report timer value = %s", cfgItem.Value)
	duration, err := time.ParseDuration(cfgItem.Value)
	if err != nil {
		return duration, err
	}
	return duration, nil
}

func (db AppDatabase) GetPropogationStatus(nodeId int64) bool {
	propogateUpstream := true //default value if it PropogateDeviceConfig does not exist in the database for this node
	cfgPropItem := PbNodeConfigItem{NodeId: nodeId, ConfigItem: ConfigItem{Key: "PropogateDeviceConfig"}}
	exists, err := cfgPropItem.Exists(db)
	if err != nil {
		db.log.Info("Error checking existence of PropogateDeviceConfig config item for the node.Error:%s", err.Error())
		return propogateUpstream
	}
	if exists {
		if err := cfgPropItem.Get(db); err != nil {
			db.log.Info("Could not recover the propogation configuration item for the root node. Error:%s", err.Error())
			return propogateUpstream
		}
		propogateUpstream = cfgPropItem.Value == "true"
	}
	return propogateUpstream
}
