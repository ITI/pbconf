package internode

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
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	change "github.com/iti/pbconf/lib/pbchange"
	database "github.com/iti/pbconf/lib/pbdatabase"
	global "github.com/iti/pbconf/lib/pbglobal"
	logging "github.com/iti/pbconf/lib/pblogger"
)

type InterNode struct {
	log logging.Logger
	db  database.AppDatabase
}

var Qmutex sync.Mutex

type Qelement struct {
	Fn_name string
	Device  *database.PbDevice
	Param   string
	Param2  string
	Cfg     *change.ChangeData
}

type UpstreamTransInfo struct {
	IpInfo        string
	TransactionId string
}

func (up UpstreamTransInfo) IP() string {
	return up.IpInfo
}
func (up UpstreamTransInfo) Transaction() string {
	return up.TransactionId
}

func NewInterNodeCommunicator(d database.AppDatabase) *InterNode {
	l, _ := logging.GetLogger("InterNode Comm")
	loglevel := logging.GetLevel("HTTP API")
	logging.SetLevel(loglevel, "InterNode Comm")
	return &InterNode{log: l, db: d}
}

/******************************* Device functions *********************************/
// (false, nil) return value means the
// function succeded. If the first return parameter is true, it should be added to queue of things to
// be retried when the connection is restored.
// Note: device here is the modified device. The deviceName is the pre-change device name, NECESSARY to find the
// device upstream
func (a *InterNode) UpdateDeviceUpstream(device database.PbDevice, deviceName string) (bool, error) {
	//is there an upstream node
	upstreamNodeIPInfo, err := a.GetUpstreamNodeIP()
	if err != nil {
		a.log.Debug("UpdateDeviceUpstream: Error getting upstream node")
		return false, err
	}
	if upstreamNodeIPInfo == nil { //no upstream node, nothing more to do return
		return false, nil
	}
	//AT THIS POINT THERE IS AN UPSTREAM NODE
	//pass upstream, get the device id upstream
	upstreamDevId, upstreamParentId, err := a.getDeviceIdsAtIP(deviceName, *upstreamNodeIPInfo)
	if err != nil {
		if strings.HasPrefix(err.Error(), global.NoConnection) || strings.HasPrefix(err.Error(), global.NoHttpOK) {
			return true, nil //PUT ON QUEUE HERE
		}
		return false, err
	}
	//reset ids to what it is upstream
	device.Id = *upstreamDevId
	device.ParentNode = upstreamParentId

	jsonStr, err := json.Marshal(device)
	if err != nil {
		a.log.Debug("UpdateDeviceUpstream: Error marshaling device content to send upstream")
		return false, err
	}
	err = a.passHttpRequest(jsonStr, *upstreamNodeIPInfo, "/device", "PATCH")
	if err != nil {
		//PUT ON QUEUE HERE
		return true, nil
	}
	return false, nil
}

//This function knows there is some node closer to device. sends request to update the device to the node below it.
// needs to return success or failure
func (a *InterNode) UpdateDeviceDownstream(device database.PbDevice, deviceName string, downstreamInfo string) bool {
	downstreamDevId, parentId, err := a.getDeviceIdsAtIP(deviceName, downstreamInfo)
	if err != nil {
		return false
	}
	//if no error, send the request to update downstream
	//reset ids to match the device ids downstream
	device.Id = *downstreamDevId
	device.ParentNode = parentId
	jsonStr, err := json.Marshal(device)
	if err != nil {
		a.log.Debug("UpdateDeviceDownstream: Error marshaling device content to send downstream")
		return false
	}
	err = a.passHttpRequest(jsonStr, downstreamInfo, "/device", "PATCH")
	if err != nil {
		return false
	}
	return true
}

// AddDeviceUpstream tries to add the device at the upstream node. (putonq=false, err=nil) return value means the
// function succeded. If the first return parameter is true, it should be added to queue of things to
// be retried when the connection is restored.
func (a *InterNode) AddDeviceUpstream(device database.PbDevice) (bool, error) {
	//is there an upstream node
	upstreamNodeIPInfo, err := a.GetUpstreamNodeIP()
	if err != nil {
		a.log.Debug("AddDeviceUpstream: Error getting upstream node")
		return false, err
	}
	if upstreamNodeIPInfo == nil { //no upstream node, nothing more to do return
		return false, nil
	}
	//AT THIS POINT THERE IS AN UPSTREAM NODE
	device.Id = 0 //reset the id, otherwise the upstream node will not process the Post request
	//set the device's parentId to this nodes' id as specified at upstream node
	rootIdUpstream, err := a.getRootIdAtIp(*upstreamNodeIPInfo)
	if err != nil {
		a.log.Debug("AddDeviceUpstream: Could not retrieve root node id at upstream node, aborting any further passing upstream")
		return true, nil
	}
	device.ParentNode = rootIdUpstream //do need to reset parent node for post and patch

	jsonStr, err := json.Marshal(device)
	if err != nil {
		a.log.Debug("AddDeviceUpstream: Error marshaling device content to send upstream")
		return false, err
	}
	err = a.passHttpRequest(jsonStr, *upstreamNodeIPInfo, "/device", "POST")
	if err != nil {
		return true, nil
	}
	return false, nil
}

// (false, nil) return value means the
// function succeded. If the first return parameter is true, it should be added to queue of things to
// be retried when the connection is restored.
func (a *InterNode) DeleteDeviceUpstream(device database.PbDevice) (bool, error) {
	//is there an upstream node
	upstreamNodeIPInfo, err := a.GetUpstreamNodeIP()
	if err != nil {
		a.log.Debug("Error getting upstream node")
		return false, err
	}
	if upstreamNodeIPInfo == nil { //no upstream node, nothing more to do return
		return false, nil
	}
	//AT THIS POINT THERE IS AN UPSTREAM NODE
	//pass upstream, get the device id upstream first
	upstreamDevId, _, err := a.getDeviceIdsAtIP(device.Name, *upstreamNodeIPInfo)
	if err != nil {
		if strings.HasPrefix(err.Error(), global.NoConnection) || strings.HasPrefix(err.Error(), global.NoHttpOK) {
			return true, nil //PUT ON QUEUE HERE
		}
		return false, err
	}
	err = a.passHttpRequest(nil, *upstreamNodeIPInfo, fmt.Sprintf("/device/%v", *upstreamDevId), "DELETE")
	if err != nil {
		//PUT ON QUEUE HERE
		return true, nil
	}
	return false, nil
}

func (a *InterNode) DeleteDeviceDownstream(device database.PbDevice, downstreamInfo string) bool {
	downstreamDevId, _, err := a.getDeviceIdsAtIP(device.Name, downstreamInfo)
	if err != nil {
		return false
	}
	err = a.passHttpRequest(nil, downstreamInfo, fmt.Sprintf("/device/%v", *downstreamDevId), "DELETE")
	if err != nil {
		return false
	}
	return true
}

// (false, nil) return value means the
// function succeded. If the first return parameter is true, it should be added to queue of things to
// be retried when the connection is restored.
func (a *InterNode) DeleteDeviceConfigItemUpstream(device database.PbDevice, cfgKey string) (bool, error) {
	//is there an upstream node
	upstreamNodeIPInfo, err := a.GetUpstreamNodeIP()
	if err != nil {
		a.log.Debug("Error getting upstream node")
		return false, err
	}
	if upstreamNodeIPInfo == nil { //no upstream node, nothing more to do return
		return false, nil
	}
	//AT THIS POINT THERE IS AN UPSTREAM NODE
	upstreamDevId, _, err := a.getDeviceIdsAtIP(device.Name, *upstreamNodeIPInfo)
	if err != nil {
		if strings.HasPrefix(err.Error(), global.NoConnection) || strings.HasPrefix(err.Error(), global.NoHttpOK) {
			return true, nil //PUT ON QUEUE HERE
		}
		return false, err
	}
	err = a.passHttpRequest(nil, *upstreamNodeIPInfo, fmt.Sprintf("/device/%v/%s", *upstreamDevId, cfgKey), "DELETE")
	if err != nil {
		//PUT ON QUEUE HERE
		return true, nil
	}
	return false, nil
}

func (a *InterNode) DeleteDeviceConfigItemDownstream(device database.PbDevice, cfgKey string, downstreamInfo string) bool {
	downstreamDevId, _, err := a.getDeviceIdsAtIP(device.Name, downstreamInfo)
	if err != nil {
		return false
	}
	err = a.passHttpRequest(nil, downstreamInfo, fmt.Sprintf("/device/%v/%s", downstreamDevId, cfgKey), "DELETE")
	if err != nil {
		return false
	}
	return true
}

// (false, nil) return value means the
// function succeded. If the first return parameter is true, it should be added to queue of things to
// be retried when the connection is restored.
func (a *InterNode) UpdateDeviceConfigUpstream(device database.PbDevice, cfg *change.ChangeData) (bool, error) {
	//is there an upstream node
	upstreamNodeIPInfo, err := a.GetUpstreamNodeIP()
	if err != nil {
		a.log.Debug("Error getting upstream node")
		return false, err
	}
	if upstreamNodeIPInfo == nil { //no upstream node, nothing more to do return
		return false, nil
	}
	//AT THIS POINT THERE IS AN UPSTREAM NODE
	jsonStr, err := json.Marshal(cfg)
	if err != nil {
		a.log.Debug("UpdateDeviceConfigUpstream: Error marshaling device content to send upstream")
		return false, err
	}
	// get this devices' info
	upstreamDevId, _, err := a.getDeviceIdsAtIP(device.Name, *upstreamNodeIPInfo)
	if err != nil {
		a.log.Debug("passDeviceConfigUpstream: Error getting device id on upstream node")
		if strings.HasPrefix(err.Error(), global.NoConnection) || strings.HasPrefix(err.Error(), global.NoHttpOK) {
			return true, nil //PUT ON QUEUE HERE
		}
		return false, err
	}

	err = a.passHttpRequest(jsonStr, *upstreamNodeIPInfo, fmt.Sprintf("/device/%v/config", *upstreamDevId), "PATCH")
	if err != nil {
		//PUT ON QUEUE HERE
		return true, nil
	}
	return false, nil
}

func (a *InterNode) UpdateDeviceConfigDownstream(device database.PbDevice, cfg *change.ChangeData, downstreamInfo string) bool {
	downstreamDevId, parentId, err := a.getDeviceIdsAtIP(device.Name, downstreamInfo)
	if err != nil {
		return false
	}
	//if no error, send the request to update downstream
	//reset ids to match the device ids downstream
	device.Id = *downstreamDevId
	device.ParentNode = parentId
	jsonStr, err := json.Marshal(cfg)
	if err != nil {
		a.log.Debug("UpdateDeviceDownstream: Error marshaling device content to send downstream")
		return false
	}
	err = a.passHttpRequest(jsonStr, downstreamInfo, fmt.Sprintf("/device/%v/config", *downstreamDevId), "PATCH")
	if err != nil {
		return false
	}
	return true

}

func (a *InterNode) UpdateDeviceMetaUpstream(device database.PbDevice, metaKey, metaValue string) (bool, error) {
	//is there an upstream node
	upstreamNodeIPInfo, err := a.GetUpstreamNodeIP()
	if err != nil {
		a.log.Debug("UpdateDeviceMetaUpstream: Error getting upstream node")
		return false, err
	}
	if upstreamNodeIPInfo == nil { //no upstream node, nothing more to do return
		return false, nil
	}
	//AT THIS POINT THERE IS AN UPSTREAM NODE
	//pass upstream, get the device id upstream
	upstreamDevId, _, err := a.getDeviceIdsAtIP(device.Name, *upstreamNodeIPInfo)
	if err != nil {
		if strings.HasPrefix(err.Error(), global.NoConnection) || strings.HasPrefix(err.Error(), global.NoHttpOK) {
			return true, nil //PUT ON QUEUE HERE
		}
		return false, err
	}

	metaItem := database.ConfigItem{Key: metaKey, Value: metaValue}
	jsonStr, err := json.Marshal(metaItem)
	if err != nil {
		a.log.Debug("UpdateDeviceMetaUpstream: Error marshaling device metadata content to send upstream")
		return false, err
	}
	err = a.passHttpRequest(jsonStr, *upstreamNodeIPInfo, fmt.Sprintf("/device/%v/meta", *upstreamDevId), "PATCH")
	if err != nil {
		//PUT ON QUEUE HERE
		return true, nil
	}
	return false, nil
}

//This function knows there is some node closer to device. sends request to update the device to the node below it.
// needs to return success or failure
func (a *InterNode) UpdateDeviceMetaDownstream(device database.PbDevice, metaKey, metaValue, downstreamInfo string) bool {
	downstreamDevId, _, err := a.getDeviceIdsAtIP(device.Name, downstreamInfo)
	if err != nil {
		return false
	}
	//if no error, send the request to update downstream
	metaItem := database.ConfigItem{Key: metaKey, Value: metaValue}
	jsonStr, err := json.Marshal(metaItem)
	if err != nil {
		a.log.Debug("UpdateDeviceDownstream: Error marshaling device metadata content to send downstream")
		return false
	}
	err = a.passHttpRequest(jsonStr, downstreamInfo, fmt.Sprintf("/device/%v/meta", *downstreamDevId), "PATCH")
	if err != nil {
		return false
	}
	return true
}
func (a *InterNode) DeleteDeviceMetaUpstream(device database.PbDevice, metaKey, metaValue string) (bool, error) {
	//is there an upstream node
	upstreamNodeIPInfo, err := a.GetUpstreamNodeIP()
	if err != nil {
		a.log.Debug("UpdateDeviceMetaUpstream: Error getting upstream node")
		return false, err
	}
	if upstreamNodeIPInfo == nil { //no upstream node, nothing more to do return
		return false, nil
	}
	//AT THIS POINT THERE IS AN UPSTREAM NODE
	//pass upstream, get the device id upstream
	upstreamDevId, _, err := a.getDeviceIdsAtIP(device.Name, *upstreamNodeIPInfo)
	if err != nil {
		if strings.HasPrefix(err.Error(), global.NoConnection) || strings.HasPrefix(err.Error(), global.NoHttpOK) {
			return true, nil //PUT ON QUEUE HERE
		}
		return false, err
	}

	metaItem := database.ConfigItem{Key: metaKey, Value: metaValue}
	jsonStr, err := json.Marshal(metaItem)
	if err != nil {
		a.log.Debug("DeleteDeviceMetaUpstream: Error marshaling device metadata content to send upstream")
		return false, err
	}
	err = a.passHttpRequest(jsonStr, *upstreamNodeIPInfo, fmt.Sprintf("/device/%v/meta", *upstreamDevId), "DELETE")
	if err != nil {
		//PUT ON QUEUE HERE
		return true, nil
	}
	return false, nil
}

//This function knows there is some node closer to device. sends request to update the device to the node below it.
// needs to return success or failure
func (a *InterNode) DeleteDeviceMetaDownstream(device database.PbDevice, metaKey, metaValue, downstreamInfo string) bool {
	downstreamDevId, _, err := a.getDeviceIdsAtIP(device.Name, downstreamInfo)
	if err != nil {
		return false
	}
	//if no error, send the request to update downstream
	metaItem := database.ConfigItem{Key: metaKey, Value: metaValue}
	jsonStr, err := json.Marshal(metaItem)
	if err != nil {
		a.log.Debug("DeleteDeviceDownstream: Error marshaling device metadata content to send downstream")
		return false
	}
	err = a.passHttpRequest(jsonStr, downstreamInfo, fmt.Sprintf("/device/%v/meta", *downstreamDevId), "DELETE")
	if err != nil {
		return false
	}
	return true
}

/********************************* Node functions ***********************************************/
// UpdateIPAddressUpstream function updates the rootNode's IP address at the upstream node with the port number
// at which the root node can be contacted. It gets the root node from the upstream node, looks for the
// IP Address config item and sends a partial node with just the modified IP Address config item to the upstream node
// with a request to PATCH the node.
func (a *InterNode) UpdateIPAddressUpstream() error {
	upstreamNodeIP, err := a.GetUpstreamNodeIP()
	if err != nil {
		return err
	}
	if upstreamNodeIP == nil {
		return nil
	}
	rootIdUpstream, err := a.getRootIdAtIp(*upstreamNodeIP)
	if err != nil {
		return err
	}

	destUrl := fmt.Sprintf("https://%s/node/%v", *upstreamNodeIP, *rootIdUpstream)
	upstream_req, err := http.NewRequest("GET", destUrl, nil)
	a.log.Debug("Sending request " + destUrl + " : GET")
	resp, err := global.Client.Do(upstream_req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		a.log.Debug("Was not able to connect upstream. Node did not respond")
		return err
	}
	var node database.PbNode
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&node); err != nil {
		a.log.Debug("Decoder error: %s", err.Error())
		return err
	}
	if node.Name != global.RootNode {
		return errors.New("Can only update root node upstream")
	}

	var ipConfig string
	for i := range node.ConfigItems {
		confItem := &node.ConfigItems[i]
		if confItem.Key == "IP Address" {
			upstreamInfo := strings.Split(confItem.Value, ":")
			ipConfig = upstreamInfo[0] + global.ApiPort
			break
		}
	}
	updatedNode := database.PbNode{Id: *rootIdUpstream, Name: node.Name, ConfigItems: []database.ConfigItem{{Key: "IP Address", Value: ipConfig}}}
	a.log.Debug("GOING TO UPDPATE UPSTREAM NODE AS %v", updatedNode)
	jsonStr, err := json.Marshal(updatedNode)
	if err != nil {
		a.log.Debug("UpdateNodeUpstream: Error marshaling device content to send upstream")
		return err
	}
	return a.passHttpRequest(jsonStr, *upstreamNodeIP, "/node", "PATCH")
}

/********************************* Policy functions ***********************************************/

func (a *InterNode) GetUpstreamNodePolicies() ([]change.ChangeData, error) {
	upstreamNodeIP, err := a.GetUpstreamNodeIP()
	if err != nil {
		return nil, err
	}
	if upstreamNodeIP == nil {
		return nil, nil
	}
	//AT THIS POINT WE ARE SURE WE HAVE AN UPSTREAM NODE
	destUrl := "https://" + *upstreamNodeIP + "/policy/all"
	req, err := http.NewRequest("GET", destUrl, nil)
	req.Header.Set("Accept-Encoding", "gzip")
	a.log.Debug("Sending request to %s", destUrl)
	resp, err := global.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		//NO NEED TO PUT ON QUEUE, on resuming the heartbeat, the node will try to update the policies when it discovers they are out of sync.
		return nil, errors.New(fmt.Sprintf(global.NoConnection+" Error:%s", err.Error()))
	}
	//upstream node responded with ok or something, check it
	if resp.StatusCode != http.StatusOK {
		a.log.Debug("GetUpstreamNodePolicies: Got response other than http.StatusOK from upstream node in response to policy get")
		a.logResponseErr(resp.Body)
		return nil, errors.New(global.NoHttpOK)
	}
	policyList := make([]change.ChangeData, 0)
	gzipReader, err := gzip.NewReader(resp.Body) //we expect a gzip compressed response
	if err != nil {
		a.log.Debug("Could not get a gzip reader for the response, error:%s", err.Error())
		return nil, err
	}
	decoder := json.NewDecoder(gzipReader)
	err = decoder.Decode(&policyList)
	if err != nil {
		a.log.Debug("GetUpstreamNodePolicies Decoder error: %s", err.Error())
		return nil, err
	}
	return policyList, nil
}

// UpdatePolicyUpstream returns error as return value
func (a *InterNode) UpdatePolicyUpstream(policy *change.ChangeData, upstreamNodeIP *string) error {
	//AT THIS POINT WE ARE SURE WE HAVE AN UPSTREAM NODE
	jsonStr, err := json.Marshal(policy)
	if err != nil {
		a.log.Debug("UpdatePolicyUpstream: Could not marshal policy to send upstream Error: %s", err.Error())
		return err
	}
	err = a.passHttpRequest(jsonStr, *upstreamNodeIP, "/policy/"+policy.Content.Object, "PUT")
	return err
}

func (a *InterNode) DeletePolicyUpstream(policyName string, upstreamNodeIP *string) error {
	err := a.passHttpRequest(nil, *upstreamNodeIP, fmt.Sprintf("/policy/%s", policyName), "DELETE")
	return err
}

/********************************* Utility functions ***********************************************/
func (a *InterNode) GetRootNode() (database.PbNode, error) {
	node := database.PbNode{Name: global.RootNode}
	err := node.GetByName(a.db)
	if err != nil {
		a.log.Debug("GetRootNode::Could not recover root node from database, Error: %s", err.Error())
	}
	return node, err
}

func (a *InterNode) GetUpstreamNodeIP() (*string, error) {
	//find the root node first
	node, err := a.GetRootNode()
	if err != nil {
		a.log.Debug("Could not get root node got error: %s", err.Error())
		return nil, err
	}

	//find the root node's upstream node if it exists
	cfgItem := database.PbNodeConfigItem{NodeId: node.Id, ConfigItem: database.ConfigItem{Key: "UpstreamNode"}}
	exists, err := cfgItem.Exists(a.db)
	if err != nil {
		a.log.Debug("GetUpstreamNodeIP: error checking existence of upstreamNodeIP: %s", err.Error())
		return nil, err
	}
	if !exists {
		a.log.Debug("GetUpstreamNodeIP: No Upstream node specified")
		return nil, nil
	}
	if err := cfgItem.Get(a.db); err != nil {
		a.log.Debug("GetUpstreamNodeIP: Could not recover the upstream node (GET), got error: %s", err.Error())
		return nil, err
	}
	upstreamNodeIP := cfgItem.Value
	if upstreamNodeIP == "" {
		return nil, nil
	}

	upstreamInfo := strings.Split(upstreamNodeIP, ":")
	if len(upstreamInfo) < 2 {
		upstreamNodeIP = upstreamInfo[0] + global.ApiPort
	}
	return &upstreamNodeIP, nil
}

// GetDownstreamNodeIP function takes as input parameter a nodeId. It looks into the database and
// returns the IP address (with port) associated with this nodeId. The IP address is
// automatically added when the node is added to this nodes' hierarchy.
func (a *InterNode) GetDownstreamNodeIP(nodeId int64) (string, error) {
	//find the root node's upstream node if it exists
	cfgItem := database.PbNodeConfigItem{NodeId: nodeId, ConfigItem: database.ConfigItem{Key: "IP Address"}}
	exists, err := cfgItem.Exists(a.db)
	if err != nil {
		a.log.Debug("GetDownstreamNodeIP: error checking existence of node IP Address config item: %s", err.Error())
		return "", err
	}
	if !exists {
		a.log.Debug("GetDownstreamNodeIP: Node with id %v does not have an IP Address", nodeId)
		return "", errors.New("Node does not have IP Address associated with it")
	}
	if err := cfgItem.Get(a.db); err != nil {
		a.log.Debug("GetDownstreamNodeIP: Could not recover the IP Address for node, got error: %s", err.Error())
		return "", err
	}
	nodeIP := cfgItem.Value
	if nodeIP == "" {
		return "", errors.New("Node IP Address is blank")
	}
	return nodeIP, nil
}

/********************************* internal functions ******************************************/
func (a *InterNode) getRootIdAtIp(upstreamNodeIP string) (*int64, error) {
	destUrl := "https://" + upstreamNodeIP + "/node"
	upstream_req, err := http.NewRequest("GET", destUrl, nil)
	a.log.Debug("Asking upstream node for all nodes")
	resp, err := global.Client.Do(upstream_req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		a.log.Debug("Was not able to ask upstream node for device id. Node did not respond")
		return nil, err
	}
	//upstream node responded with ok or something, check it
	if resp.StatusCode != http.StatusOK {
		a.log.Debug("Got response other than http.StatusOK from upstream node in response getting all nodes")
		a.logResponseErr(resp.Body)
		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)
	var nodeList []database.PbNode
	err = decoder.Decode(&nodeList)
	if err != nil {
		a.log.Debug("Decoder error: %s", err.Error())
		return nil, err
	}

	for _, node := range nodeList {
		if node.Name == global.RootNode {
			return &node.Id, nil
		}
	}
	return nil, errors.New("Could not find specified node upstream")
}

// passHttpRequest passes a http request to the input ipAddress+route. It returns the global error NoConnection or
// NoHttpOK as errors or nil(no errors, request successfully sent). This function CLOSES the response body, so it is
// not appropriate to call this function where the caller sends a body
func (a *InterNode) passHttpRequest(jsonStr []byte, ipAddress string, route string, httpVerb string) error {
	destUrl := "https://" + ipAddress + route
	req, err := http.NewRequest(httpVerb, destUrl, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	a.log.Debug("Sending request %s to %s", httpVerb, destUrl)
	resp, err := global.Client.Do(req)
	a.log.Debug("AFTER global.Client.Do****")
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		a.log.Debug("Was not able to pass request to %v. Node did not respond", ipAddress)
		return errors.New(fmt.Sprintf(global.NoConnection+" Error:%s", err.Error()))
	}
	//downstream node responded with  NOT ok or something, check it
	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
		a.log.Debug("Got response other than http.StatusOK from node with IP address: %s", ipAddress)
		a.logResponseErr(resp.Body)
		return errors.New(global.NoHttpOK)
	}
	return nil
}

func (a *InterNode) getDeviceAtDownstreamNode(deviceName, downStreamInfo string) (*database.PbDevice, error) {
	destUrl := "https://" + downStreamInfo + "/device"
	req, err := http.NewRequest("GET", destUrl, nil)
	a.log.Debug("Asking downstream node for all its devices")
	resp, err := global.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		a.log.Debug("getDeviceAtDownstreamNode: Was not able to ask downstream node for device. Node did not respond Error: %s", err.Error())
		return nil, err
	}
	//upstream node responded with ok or something, check it
	if resp.StatusCode != http.StatusOK {
		a.log.Debug("getDeviceAtDownstreamNode: Got response other than http.StatusOK from downstream node in response to device detail")
		a.logResponseErr(resp.Body)
		return nil, err
	}
	decoder := json.NewDecoder(resp.Body)
	var deviceList []database.PbDevice
	err = decoder.Decode(&deviceList)
	if err != nil {
		a.log.Debug("getDeviceAtDownstreamNode Decoder error: %s", err.Error())
		return nil, err
	}
	for _, dev := range deviceList {
		if dev.Name == deviceName {
			a.log.Debug("Got device from downstream node")
			return &dev, nil
		}
	}
	return nil, errors.New("Could not find specified device downstream")
}

//getDeviceIdsAtIP returns the device at the specified IP, it returns
// the parentNode and the Id of the device.
func (a *InterNode) getDeviceIdsAtIP(deviceName, ipInfo string) (*int64, *int64, error) {
	destUrl := "https://" + ipInfo + "/device"
	req, err := http.NewRequest("GET", destUrl, nil)
	a.log.Debug("Asking node with IP %s for all its devices", ipInfo)
	resp, err := global.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		a.log.Debug("getDeviceIdsAtIP: Was not able to ask node for device info. Node did not respond Error: %s", err.Error())
		return nil, nil, errors.New(fmt.Sprintf(global.NoConnection+" Error:%s", err.Error()))
	}
	//upstream node responded with NOT ok or something, check it
	if resp.StatusCode != http.StatusOK {
		a.log.Debug("getDeviceIdsAtIP: Got response other than http.StatusOK from node in response to device detail")
		return nil, nil, errors.New(fmt.Sprintf(global.NoHttpOK+" Error:%s", err.Error()))
	}
	decoder := json.NewDecoder(resp.Body)
	var deviceList []database.PbDevice
	err = decoder.Decode(&deviceList)
	if err != nil {
		a.log.Debug("getDeviceIdsAtIP Decoder error: %s", err.Error())
		return nil, nil, err
	}
	for _, dev := range deviceList {
		if dev.Name == deviceName {
			a.log.Debug("Got device from node")
			return &dev.Id, dev.ParentNode, nil
		}
	}
	return nil, nil, errors.New(fmt.Sprintf("Could not find specified device at node with IP %s", ipInfo))
}

func (a *InterNode) logResponseErr(body io.ReadCloser) {
	//now try to figure out what the error reported was
	var logmsg logging.LogMsg
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&logmsg)
	if err == nil { //if we were able to decode it, log it
		a.log.Log(logmsg.Level, logmsg.SrcNode+" Node Error: "+logmsg.Message)
	}
}
