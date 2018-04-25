package pblogger

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
	"crypto/tls"
	"errors"
	"sync"

	logging "github.com/iti/go-logging"
	config "github.com/iti/pbconf/lib/pbconfig"
	"text/template"

	"bytes"
	"fmt"
	"net/smtp"
)

func init() {
	AlarmRegister("email", NewEmailAlarm)
}

type emailMessage struct {
	From, To, Msg string
}

type EmailAlarm struct {
	FromAddr string
	ToAddr   string
	Server   string
	Template string

	TLSConfig *tls.Config

	UseTLS          bool
	IgnoreCertError bool
	ForceAuth       bool

	Password string
}

func NewEmailAlarm(cfg *config.Config) AlarmDest {
	return &EmailAlarm{
		FromAddr: cfg.Global.EmailFrom,
		ToAddr:   cfg.Global.EmailTo,
		Server:   cfg.Global.EmailServer,
		Template: cfg.Global.EmailTemplate,
	}
}

func (alarm *EmailAlarm) Emit(rec *logging.Record, wg *sync.WaitGroup) error {

	//compatibility with prerefactor code
	b := alarm
	msgBuf := new(bytes.Buffer)
	TLSon := false

	t, _ := template.New("Email").Parse(alarm.Template)
	t.Execute(msgBuf, emailMessage{b.FromAddr, b.ToAddr, rec.Message()})

	// Connect
	c, err := smtp.Dial(b.Server)
	if err != nil {
		return alarm.err(err, wg)
	}
	defer c.Close()

	// HELO - Probably should be IP of node
	if err = c.Hello("PBCONF"); err != nil {
		return alarm.err(err, wg)
	}

	if b.TLSConfig != nil || b.UseTLS {
		// Check Server
		if ok, _ := c.Extension("STARTTLS"); ok {
			if b.TLSConfig == nil {
				b.TLSConfig = &tls.Config{ServerName: b.Server}
			}
			if b.IgnoreCertError {
				b.TLSConfig.InsecureSkipVerify = true
			}
			if err = c.StartTLS(b.TLSConfig); err == nil {
				TLSon = true
			}
		} else {
			// Server doesn't support TLS
			if b.UseTLS {
				return alarm.err(errors.New("UseTLS set and Server doesn't support TLS"), wg)
			}
		}
	}

	if b.Password != "" || b.ForceAuth {
		if ok, _ := c.Extension("AUTH"); ok {
			_ = TLSon
			// Auth bits
		} else {
			if b.ForceAuth {
				return alarm.err(errors.New("ForceAuth set but server doesn't support AUTH"), wg)
			}
		}
	}

	if err := c.Mail(b.FromAddr); err != nil {
		fmt.Printf("%s", err.Error())
		return alarm.err(err, wg)
	}

	if err := c.Rcpt(b.ToAddr); err != nil {
		fmt.Printf("%s", err.Error())
		return alarm.err(err, wg)
	}

	wc, err := c.Data()
	if err != nil {
		fmt.Printf("%s", err.Error())
		return alarm.err(err, wg)
	}

	_, err = fmt.Fprintf(wc, msgBuf.String())
	if err != nil {
		fmt.Printf("%s", err.Error())
		return alarm.err(err, wg)
	}

	err = wc.Close()
	if err != nil {
		fmt.Printf("%s", err.Error())
		return alarm.err(err, wg)
	}

	err = c.Quit()
	if err != nil {
		fmt.Printf("%s", err.Error())
		return alarm.err(err, wg)
	}

	return alarm.err(nil, wg)
}

func (alarm *EmailAlarm) err(err error, wg *sync.WaitGroup) error {
	wg.Done()
	return err
}
