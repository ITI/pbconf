package pbchange

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

var activeTransactions map[string]*Transaction

func init() {
	activeTransactions = make(map[string]*Transaction, 0)
}

func (engine *CMEngine) BeginTransaction(data *ChangeData, message string) (string, error) {

	log.Debug("BeginTransaction()")

	var id string
	if data.TransactionID != "" {
		if _, ok := activeTransactions[data.TransactionID]; ok {
			return "", NewCMError("Transaction ID Conflict")
		}
		id = data.TransactionID
	} else {
		var err error
		id, err = engine.getUUID(5, func(id string) bool {
			_, ok := activeTransactions[id]
			return !ok
		})
		if err != nil {
			log.Debug("Could not create a transaction ID")
			return "", NewCMCollisionError()
		}
	}

	activeTransactions[id] = &Transaction{Status: INITIALIZING, Ctype: data.ObjectType}

	_, ckerr := engine.getGitDir(data.ObjectType)
	if ckerr != nil {
		log.Debug("Repo Missing")
		if mrerr := engine.MakeRepo(data.ObjectType); mrerr != nil {
			log.Debug("Failed to make repo: %s", mrerr.Error())
			return "", mrerr
		}
	}

	// Create the branch, we do not need a lock here
	if _, err := engine.run(data.ObjectType, "branch", id); err != nil {
		log.Warning("Failed to create transaction branch")
		return "", NewCMTransactionError(id)
	}

	// does its own locking
	_, err := engine.VersionObject(data, message, id)
	if err != nil {
		activeTransactions[id].Status = FAILED
		return "", err
	}

	activeTransactions[id].Status = ACTIVE

	return id, nil
}

func (engine *CMEngine) FinalizeTransaction(cbdata *ChangeData, status ...TransactionStatus) error {

	var newStatus TransactionStatus
	var rerr error = nil

	id := cbdata.TransactionID

	if len(status) == 0 {
		newStatus = COMPLETE
	} else {
		newStatus = status[0]
	}

	if _, ok := activeTransactions[id]; ok {
		activeTransactions[id].Status = newStatus
	} else {
		rerr = NewCMError("Unknown Transaction")
	}

	// Make sure the CBdata we send has the correct trans ID
	cdata := *cbdata
	engine.runCommitCBs(cdata.ObjectType, &cdata)

	log.Debug("Current Transactions: %v", activeTransactions)
	return rerr
}
