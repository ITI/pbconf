package main

import (
	"flag"
	"fmt"
	"github.com/btcsuite/golangcrypto/bcrypt"
)

func main() {
	uname := flag.String("u", "", "Username")
	pass := flag.String("p", "", "Password")
	email := flag.String("e", "", "Email")
	role := flag.String("r", "user", "Role")

	flag.Parse()

	switch {
	case *uname == "":
		panic("Missing username")
	case *pass == "":
		panic("Missing password")
	case *email == "":
		panic("Missing email")
	}

	hp, err := bcrypt.GenerateFromPassword([]byte(*pass), 10)
	if err != nil {
		panic(err)
	}

	fmt.Printf("INSERT INTO Users (name, passwordHash, email, role) values (\"%s\", \"%s\", \"%s\", \"%s\");\n", *uname, string(hp), *email, *role)

}
