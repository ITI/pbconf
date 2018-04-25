package pbauth

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
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"time"

	"github.com/btcsuite/golangcrypto/bcrypt"
	jwt "github.com/dgrijalva/jwt-go"
	database "github.com/iti/pbconf/lib/pbdatabase"
	logging "github.com/iti/pbconf/lib/pblogger"
)

type Auth struct {
	log    logging.Logger
	db     database.AppDatabase
	rsaKey *rsa.PrivateKey
}

type UserClaim struct {
	Name string
	Role string
}

func NewAuthHandler(loglevel string, d database.AppDatabase) *Auth {
	l, _ := logging.GetLogger("Http Authentication")
	logging.SetLevel(loglevel, "Http Authentication")
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		l.Fatal(err.Error())
	}
	return &Auth{log: l, db: d, rsaKey: rsaKey}
}

func (a *Auth) RsaPublicKey() *rsa.PublicKey {
	return &a.rsaKey.PublicKey
}

func (a *Auth) SaveNewUser(user, password string) error {
	hashedPass := a.encryptPassword(password)
	pbUser := database.PbUser{Name: user, Password: hashedPass, Role: "admin"}
	err := pbUser.Create(a.db)
	if err != nil {
		a.log.Debug("Error creating user in the database")
	}
	return err
}

func (a *Auth) UpdateUser(user, oldpassword, newpassword string) error {
	pbUser := database.PbUser{Name: user}
	if err := pbUser.GetByName(a.db); err != nil {
		a.log.Debug("Error while retrieving user info from database in UpdateUser()")
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(pbUser.Password), []byte(oldpassword)); err != nil {
		a.log.Debug("Password not updated. The original password is not a match for the stored password")
		return errors.New("Passwords do not match. Not updated")
	}

	updatedUser := database.PbUser{Id: pbUser.Id, Name: user, Password: a.encryptPassword(newpassword), Role: pbUser.Role}
	if err := updatedUser.CreateOrUpdate(a.db); err != nil {
		a.log.Debug("Error updating user information")
		return err
	}
	return nil
}

//GetUserInfoFromStore function is the plugin that returns the hashed password from wherever we have stored it.
func (a *Auth) GetUserInfoFromStore(user string) *database.PbUser {
	pbUser := database.PbUser{Name: user}
	if err := pbUser.GetByName(a.db); err != nil {
		a.log.Debug("GetUserInfoFromStore error while retrieving user info from database, Error:%s", err.Error())
		return nil
	}
	return &pbUser
}

func (a *Auth) AuthenticateUser(user, pass string) (*database.PbUser, error) {
	usrInfo := a.GetUserInfoFromStore(user)
	if usrInfo == nil {
		return nil, errors.New("Could not retrieve user info from db")
	}
	if bcrypt.CompareHashAndPassword([]byte(usrInfo.Password), []byte(pass)) != nil {
		return nil, errors.New("Password does not match")
	}
	return usrInfo, nil
}

func (a *Auth) GenerateToken(usrInfo *database.PbUser, expiryTime int) (string, error) {
	t := jwt.New(jwt.GetSigningMethod("RS256"))
	//set our claims
	t.Claims["UserInfo"] = UserClaim{usrInfo.Name, usrInfo.Role}
	//set the expire time
	t.Claims["exp"] = time.Now().Add(time.Minute * time.Duration(expiryTime)).Unix()
	tokenString, err := t.SignedString(a.rsaKey)
	return tokenString, err
}

func (a *Auth) RenewToken(token *jwt.Token, expiryTime int) (string, error) {
	token.Claims["exp"] = time.Now().Add(time.Minute * time.Duration(expiryTime)).Unix()
	tokenString, err := token.SignedString(a.rsaKey)
	return tokenString, err
}

//encryptPassword function encrypts the plain text password with the chosen scheme.
// Internally the bcrypt.GenerateFromPassword salts and hashes the plain password.
func (a *Auth) encryptPassword(plainPass string) string {
	// Hashing the password with the cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPass), 10)
	if err != nil {
		a.log.Debug("Generating hashed password using bcrypt has failed.")
		panic(err)
	}
	return string(hashedPassword[:])
}
