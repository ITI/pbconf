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

import (
	"errors"
	"fmt"
)

// Generaic CME Error
type CMError struct {
	error
}

func NewCMError(s string) error {
	return CMError{
		error: errors.New(s),
	}
}

type CMInitError struct {
	error
}

func NewCMInitError() error {
	return CMInitError{
		error: errors.New("Change Management Engine not Initialized"),
	}
}

type CMCommunicationError struct {
	error
	Stderror string
}

func NewCMCommunicationError(o string, e error) error {
	return CMCommunicationError{
		e, o,
	}
}

// Missing Repository
type CMNoRepoError struct {
	error
}

func NewCMNoRepoError(s string) error {
	return CMNoRepoError{
		error: errors.New(fmt.Sprintf("Repository %s does not exist", s)),
	}
}

// Metadata Key missing
type CMMetaNoKeyError struct {
	error
}

func NewCMMetaNoKeyError(s string) error {
	return CMMetaNoKeyError{
		error: errors.New(fmt.Sprintf("Metadata Key %s not found", s)),
	}
}

type CMCollisionError struct {
	error
}

func NewCMCollisionError() error {
	return CMCollisionError{
		error: errors.New("To many collisions while generating transaction ID"),
	}
}

type CMTransactionError struct {
	error
}

func NewCMTransactionError(s string) error {
	return CMTransactionError{
		error: errors.New(fmt.Sprintf("Failed to create transaction branch: %s", s)),
	}
}

type CMIDError struct {
	error
}

func NewCMIDError() error {
	return CMIDError{
		error: errors.New("Failed to find an acceptable UUID"),
	}
}
