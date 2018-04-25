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

type PbUser struct {
	Id       int64
	Name     string `validate:"nonzero"`
	Password string
	Email    string
	Role     string
}

func (n *PbUser) Get(db AppDatabase) error {
	err := db.QueryRow("SELECT name, passwordHash, email, role FROM Users WHERE id=?", n.Id).Scan(&n.Name, &n.Password, &n.Email, &n.Role)
	if err != nil {
		db.log.Debug("GetByName User error: %s", err.Error())
	}
	return err
}

func (n *PbUser) GetByName(db AppDatabase) error {
	err := db.QueryRow("SELECT id, passwordHash, email, role FROM Users WHERE name=?", n.Name).Scan(&n.Id, &n.Password, &n.Email, &n.Role)
	if err != nil {
		db.log.Debug("GetByName User error: %s", err.Error())
	}
	return err
}

func (n *PbUser) Create(db AppDatabase) error {
	res, err := db.Exec("INSERT INTO Users VALUES(?, ?, ?, ?, ?)", nil, n.Name, n.Password, n.Email, n.Role)
	if err != nil {
		db.log.Info("Create User Error: %s", err.Error())
	}

	id, err := res.LastInsertId()
	if err != nil {
		db.log.Info("Create User Error, while recovering id: %s", err.Error())
		return err
	}
	n.Id = id
	return nil
}

func (n *PbUser) CreateOrUpdate(db AppDatabase) error {
	res, err := db.Exec("INSERT OR REPLACE INTO Users VALUES(?, ?, ?, ?, ?)", n.Id, n.Name, n.Password, n.Email, n.Role)
	if err != nil {
		db.log.Info("CreateOrUpdateUser Error: %s", err.Error())
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		db.log.Info("CreateOrUpdateUser Error: %s", err.Error())
		return err
	}
	n.Id = id
	return nil
}

func (n *PbUser) Delete(db AppDatabase) error {
	_, err := db.Exec("DELETE FROM Users WHERE id=?", n.Id)
	if err != nil {
		db.log.Info("DeleteUser Error: %s", err.Error())
	}
	return err
}
