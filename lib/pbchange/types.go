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
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	logging "github.com/iti/pbconf/lib/pblogger"
)

//go:generate bash gen.sh
type CMType int

const (
	DEVICE   CMType = iota // A Device
	POLICY                 // A Policy
	QUERY                  // A Report Defination
	REPORT                 // The results of running a report
	ONTOLOGY               // Ontology stub that is later compined with policy
	NONE
)

func StringToCMType(s string) CMType {
	switch s {
	case "DEVICE":
		return DEVICE
	case "POLICY":
		return POLICY
	case "QUERY":
		return QUERY
	case "REPORT":
		return REPORT
	}

	return NONE
}

type TransactionStatus int

const (
	INITIALIZING TransactionStatus = iota
	ACTIVE
	COMPLETE
	FAILED
	CLEANED
)

type Transaction struct {
	Status TransactionStatus
	Ctype  CMType
}

type CMEngine struct {
	Repopath    string
	log         logging.Logger
	binpath     string
	commitCBs   []*cbStore
	packRcvdCBs []*cbStore

	// Remove
	UploadPack  bool
	ReceivePack bool

	guard    sync.Mutex
	metalock sync.Mutex
}

type CMAuthor struct {
	Name  string
	Email string
	When  time.Time
}

type CMContent struct {
	Files  map[string][]byte
	Object string
}

func (cm *CMContent) MarshalJSON() ([]byte, error) {
	jsonStr := `{"Files":{`
	var parts []string
	for key, val := range cm.Files {
		encodedStr, err := UrlEncoded(string(val))
		if err != nil {
			return nil, err
		}
		parts = append(parts, `"`+key+`":"`+encodedStr+`"`)
	}
	jsonStr += strings.Join(parts, ",")
	jsonStr += `},`
	jsonStr += `"Object":"` + cm.Object + `"`
	jsonStr += `}`

	return []byte(jsonStr), nil
}

func UrlEncoded(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (cm *CMContent) UnmarshalJSON(bytes []byte) error {
	cm.Files = make(map[string][]byte)
	strFiles := make(map[string]string) //use string instead of []byte to unmarshal first

	var objmap map[string]*json.RawMessage
	err := json.Unmarshal(bytes, &objmap)
	if err != nil {
		return err
	}
	err = json.Unmarshal(*objmap["Files"], &strFiles)
	if err != nil {
		return err
	}
	for key, val := range strFiles {
		decodedStr, err := url.QueryUnescape(val)
		if err != nil {
			return err
		}
		cm.Files[key] = []byte(decodedStr) //now cast to our desired []byte
	}

	err = json.Unmarshal(*objmap["Object"], &cm.Object)
	if err != nil {
		return err
	}
	return nil
}

func NewCMContent(object string) *CMContent {
	c := CMContent{Object: object}
	c.Files = make(map[string][]byte, 0)
	return &c
}

type ChangeData struct {
	ObjectType    CMType
	Content       *CMContent
	Author        *CMAuthor
	CommitID      string
	TransactionID string
	SrcNode       string
	Log           *LogLine
}

type LogLine struct {
	Time           time.Time
	Id             string
	Author         string
	AuthorEmail    string
	Committer      string
	CommitterEmail string
	Message        string
}

func (l *LogLine) String() string {
	return fmt.Sprintf("%s::<%s> Author: %s <%s> Signed off by: %s <%s> %s", l.Time.Format(time.RFC3339), l.Id, l.Author, l.AuthorEmail, l.Committer, l.CommitterEmail, l.Message)
}

type checkUUID func(string) bool

type Upstream interface {
	IP() string
	Transaction() string
}
