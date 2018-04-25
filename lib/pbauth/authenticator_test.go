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

// This file contains the testing for functions of PbUser database functionality.
// As a part of this testing, the database is NOT mocked, but instead a test
// database is created for every test.
package pbauth

import (
	"github.com/btcsuite/golangcrypto/bcrypt"
	"os"
	"testing"

	"fmt"
	config "github.com/iti/pbconf/lib/pbconfig"
	"github.com/iti/pbconf/lib/pbdatabase"
	logging "github.com/iti/pbconf/lib/pblogger"
)

func begin(t *testing.T, name string) {
	fmt.Printf("##################### Begin %s #####################\n", name)
}

func end(t *testing.T, name string) {
	fmt.Printf("###################### End %s ######################\n", name)
}

func setup() *config.Config {
	cfg := new(config.Config)
	cfg.Global.LogLevel = "DEBUG"

	return cfg
}

var logLevel = "DEBUG"

func setupDB(t *testing.T, dbFile string) pbdatabase.AppDatabase {
	cfg := setup()
	logging.InitLogger("", cfg, "")
	os.Remove(dbFile)
	dbHandle := pbdatabase.Open(dbFile, logLevel)
	if dbHandle.Ping() != nil {
		t.Error("Could not create test database file")
	}
	dbHandle.LoadSchema()
	return dbHandle
}

func setupAuthHandler(dbHandle pbdatabase.AppDatabase) *Auth {
	authHandler := NewAuthHandler(logLevel, dbHandle)
	return authHandler
}

func TestSaveNewUserHandler(t *testing.T) {
	begin(t, "TestSaveNewUserHandler")
	defer end(t, "TestSaveNewUserHandler")

	dbFile := "test_newuser.db"
	dbHandle := setupDB(t, dbFile)
	defer dbHandle.Close()

	authHandler := setupAuthHandler(dbHandle)
	err := authHandler.SaveNewUser("henry", "myblippityPass3$^h")
	if err != nil {
		t.Error(fmt.Sprintf("SaveNew User threw error: %s", err.Error()))
	}

	user := pbdatabase.PbUser{Name: "henry"}
	if err := user.GetByName(dbHandle); err != nil {
		t.Error(err.Error())
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("myblippityPass3$^h")); err != nil {
		t.Error(err.Error())
	}

	//cleanup
	os.Remove(dbFile)
}

func TestUpdateUserPassword(t *testing.T) {
	begin(t, "TestUpdateUserPassword")
	defer end(t, "TestUpdateUserPassword")

	dbFile := "test_updateuser.db"
	dbHandle := setupDB(t, dbFile)
	defer dbHandle.Close()

	authHandler := setupAuthHandler(dbHandle)

	err := authHandler.SaveNewUser("henry", "myblippityPass3$^h")
	if err != nil {
		t.Error(fmt.Sprintf("SaveNew User threw error: %s", err.Error()))
	}

	authHandler.UpdateUser("henry", "myblippityPass3$^h", "re5df#d%je9s$@5")
	//check the update went through
	user := pbdatabase.PbUser{Name: "henry"}
	if err := user.GetByName(dbHandle); err != nil {
		t.Error(err.Error())
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("re5df#d%je9s$@5")) != nil {
		t.Error("Password was not updated successfully")
	}
	//test2: give wrong original password, update user should fail
	err = authHandler.UpdateUser("henry", "re5d#d%je9s$@5", "uh-oh")
	if err == nil {
		t.Error("Password should not have updated successfully with wrong original password.")
	}

	//cleanup
	os.Remove(dbFile)
}
