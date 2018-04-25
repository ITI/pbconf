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

/* Taken from https://github.com/asim/git-http-backend/

The down -
Unfortunatly the format and validity og the pack file format is a bit tricky
to sort.  Sadly this means that the best route (for the time being) to full
smart-http support is to use the git command to provide pack building, and
if that's the case, then why not just let the git commnd provide the entire
interaction?  This should be readdressed in the furture, as a pure Go
implementation is going to better serve the needs of the application and the
application users.
--jmj

Original license:
Copyright (c) 2013 Asim Aslam <asim@aslam.me>

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
'Software'), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

*/

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	Method  string
	Handler func(handlerReq)
	Rpc     string
}

type handlerReq struct {
	w           http.ResponseWriter
	r           *http.Request
	Rpc         string
	Dir         string
	File        string
	Ctype       CMType
	Transaction string
	SrcNode     string
}

func (e *CMEngine) RequestHandler() http.HandlerFunc {
	var services = map[string]Service{
		"(.*?)/git-upload-pack$":                       Service{"POST", e.serviceRpc, "upload-pack"},
		"(.*?)/git-receive-pack$":                      Service{"POST", e.serviceRpc, "receive-pack"},
		"(.*?)/info/refs$":                             Service{"GET", e.getInfoRefs, ""},
		"(.*?)/HEAD$":                                  Service{"GET", e.getTextFile, ""},
		"(.*?)/objects/info/alternates$":               Service{"GET", e.getTextFile, ""},
		"(.*?)/objects/info/http-alternates$":          Service{"GET", e.getTextFile, ""},
		"(.*?)/objects/info/packs$":                    Service{"GET", e.getInfoPacks, ""},
		"(.*?)/objects/info/[^/]*$":                    Service{"GET", e.getTextFile, ""},
		"(.*?)/objects/[0-9a-f]{2}/[0-9a-f]{38}$":      Service{"GET", e.getLooseObject, ""},
		"(.*?)/objects/pack/pack-[0-9a-f]{40}\\.pack$": Service{"GET", e.getPackFile, ""},
		"(.*?)/objects/pack/pack-[0-9a-f]{40}\\.idx$":  Service{"GET", e.getIdxFile, ""},
	}

	// Request handling function

	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("%s %s %s %s", r.RemoteAddr, r.Method, r.URL.Path, r.Proto)
		for match, service := range services {
			re, err := regexp.Compile(match)
			if err != nil {
				log.Warning(err.Error())
			}

			if m := re.FindStringSubmatch(r.URL.Path); m != nil {
				if service.Method != r.Method {
					e.renderMethodNotAllowed(w, r)
					return
				}

				var ctype CMType
				switch {
				case strings.HasPrefix(r.URL.Path, "/device/"):
					ctype = DEVICE
				case strings.HasPrefix(r.URL.Path, "/policy/"):
					ctype = POLICY
				default:
					e.renderNotFound(w)
					return
				}

				rpc := service.Rpc
				file := strings.Replace(r.URL.Path, m[1], "", 1)

				// On push, we need to make sure the repo exists
				log.Debug("Checking Repo")
				dir, err := e.getGitDir(ctype)
				if err != nil {
					log.Debug("Repo Missing")
					if mrerr := e.MakeRepo(ctype); mrerr != nil {
						log.Debug("Failed to make repo: %s", mrerr.Error())
						e.renderNotFound(w)
						return
					} else {
						dir, err = e.getGitDir(ctype)
					}
				}

				if err != nil {
					log.Info(err.Error())
					e.renderNotFound(w)
					return
				}

				// Extract the transaction ID and src node if present
				urlTokens := strings.Split(m[1], "/")
				transID := ""
				srcNode := ""
				for _, tok := range urlTokens {
					if strings.HasPrefix(tok, "src") {
						srcNode = strings.TrimPrefix(tok, "src=")
					} else if strings.HasPrefix(tok, "trans") {
						transID = strings.TrimPrefix(tok, "trans=")
					}
				}
				log.Debug("File: %s", file)
				hr := handlerReq{
					w:           w,
					r:           r,
					Rpc:         rpc,
					Dir:         dir,
					File:        file,
					Ctype:       ctype,
					Transaction: transID,
					SrcNode:     srcNode,
				}
				service.Handler(hr)
				return
			}
		}
		e.renderNotFound(w)
		return
	}
}

// Actual command handling functions

func (e *CMEngine) serviceRpc(hr handlerReq) {
	w, r, rpc, dir, src := hr.w, hr.r, hr.Rpc, hr.Dir, hr.SrcNode
	access := e.hasAccess(r, dir, rpc, true)

	if access == false {
		e.renderNoAccess(w)
		return
	}

	input, _ := ioutil.ReadAll(r.Body)

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-result", rpc))
	w.WriteHeader(http.StatusOK)

	args := []string{rpc, "--stateless-rpc", dir}
	cmd := exec.Command(e.binpath, args...)
	cmd.Dir = dir
	in, err := cmd.StdinPipe()
	if err != nil {
		log.Warning(err.Error())
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Warning(err.Error())
	}

	err = cmd.Start()
	if err != nil {
		log.Warning(err.Error())
	}

	in.Write(input)
	io.Copy(w, stdout)
	cmd.Wait()
	if rpc == "receive-pack" {
		cbdata := ChangeData{TransactionID: hr.Transaction, ObjectType: hr.Ctype, SrcNode: src}
		if hr.Transaction != "" {
			log.Debug("Clearing Transaction %v", hr.Transaction)
			if err := e.FinalizeTransaction(&cbdata); err != nil {
				log.Warning("Failed to finalize transaction: %v", err)
			}
		}
		e.runPackRcvCBs(cbdata.ObjectType, &cbdata)

	}

	log.Debug("calling reset()")
	e.reset(hr.Ctype)
}

func (e *CMEngine) getInfoRefs(hr handlerReq) {
	log.Debug("getInfoRefs()")
	w, r, dir := hr.w, hr.r, hr.Dir
	service_name := e.getServiceType(r)
	log.Debug("Service: %s\n", service_name)
	access := e.hasAccess(r, dir, service_name, false)

	if access {
		args := []string{service_name, "--stateless-rpc", "--advertise-refs", "."}
		refs := e.gitCommand(dir, args...)

		log.Debug("Here")
		e.hdrNocache(w)
		w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-advertisement", service_name))
		w.WriteHeader(http.StatusOK)
		w.Write(e.packetWrite("# service=git-" + service_name + "\n"))
		w.Write(e.packetFlush())
		w.Write(refs)
	} else {
		e.updateServerInfo(dir)
		e.hdrNocache(w)
		e.sendFile("text/plain; charset=utf-8", hr)
	}
}

func (e *CMEngine) getInfoPacks(hr handlerReq) {
	e.hdrCacheForever(hr.w)
	e.sendFile("text/plain; charset=utf-8", hr)
}

func (e *CMEngine) getLooseObject(hr handlerReq) {
	e.hdrCacheForever(hr.w)
	e.sendFile("application/x-git-loose-object", hr)
}

func (e *CMEngine) getPackFile(hr handlerReq) {
	e.hdrCacheForever(hr.w)
	e.sendFile("application/x-git-packed-objects", hr)
}

func (e *CMEngine) getIdxFile(hr handlerReq) {
	e.hdrCacheForever(hr.w)
	e.sendFile("application/x-git-packed-objects-toc", hr)
}

func (e *CMEngine) getTextFile(hr handlerReq) {
	e.hdrNocache(hr.w)
	e.sendFile("text/plain", hr)
}

// Logic helping functions

func (e *CMEngine) sendFile(content_type string, hr handlerReq) {
	w, r := hr.w, hr.r
	req_file := path.Join(hr.Dir, hr.File)

	f, err := os.Stat(req_file)
	if os.IsNotExist(err) {
		log.Warning(err.Error())
		e.renderNotFound(w)
		return
	}

	w.Header().Set("Content-Type", content_type)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", f.Size()))
	w.Header().Set("Last-Modified", f.ModTime().Format(http.TimeFormat))
	http.ServeFile(w, r, req_file)
}

func (e *CMEngine) getServiceType(r *http.Request) string {
	service_type := r.FormValue("service")

	if s := strings.HasPrefix(service_type, "git-"); !s {
		return ""
	}

	return strings.Replace(service_type, "git-", "", 1)
}

func (e *CMEngine) hasAccess(r *http.Request, dir string, rpc string, check_content_type bool) bool {
	if check_content_type {
		if r.Header.Get("Content-Type") != fmt.Sprintf("application/x-git-%s-request", rpc) {
			return false
		}
	}

	if !(rpc == "upload-pack" || rpc == "receive-pack") {
		return false
	}
	if rpc == "receive-pack" {
		return e.ReceivePack
	}
	if rpc == "upload-pack" {
		return e.UploadPack
	}

	return e.getConfigSetting(rpc, dir)
}

func (e *CMEngine) getConfigSetting(service_name string, dir string) bool {
	service_name = strings.Replace(service_name, "-", "", -1)
	setting := e.getGitConfig("http."+service_name, dir)

	if service_name == "uploadpack" {
		return setting != "false"
	}

	return setting == "true"
}

func (e *CMEngine) getGitConfig(config_name string, dir string) string {
	args := []string{"config", config_name}
	out := string(e.gitCommand(dir, args...))
	return out[0 : len(out)-1]
}

func (e *CMEngine) updateServerInfo(dir string) []byte {
	args := []string{"update-server-info"}
	return e.gitCommand(dir, args...)
}

func (e *CMEngine) gitCommand(dir string, args ...string) []byte {
	command := exec.Command(e.binpath, args...)
	command.Dir = dir
	out, err := command.Output()

	if err != nil {
		log.Warning(err.Error())
	}

	return out
}

// HTTP error response handling functions

func (e *CMEngine) renderMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	if r.Proto == "HTTP/1.1" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method Not Allowed"))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	}
}

func (e *CMEngine) renderNotFound(w http.ResponseWriter) {
	log.Debug("not found")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
}

func (e *CMEngine) renderNoAccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("Forbidden"))
}

// Packet-line handling function

func (e *CMEngine) packetFlush() []byte {
	return []byte("0000")
}

func (e *CMEngine) packetWrite(str string) []byte {
	s := strconv.FormatInt(int64(len(str)+4), 16)

	if len(s)%4 != 0 {
		s = strings.Repeat("0", 4-len(s)%4) + s
	}

	return []byte(s + str)
}

// Header writing functions

func (e *CMEngine) hdrNocache(w http.ResponseWriter) {
	w.Header().Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
}

func (e *CMEngine) hdrCacheForever(w http.ResponseWriter) {
	now := time.Now().Unix()
	expires := now + 31536000
	w.Header().Set("Date", fmt.Sprintf("%d", now))
	w.Header().Set("Expires", fmt.Sprintf("%d", expires))
	w.Header().Set("Cache-Control", "public, max-age=31536000")
}
