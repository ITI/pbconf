package main

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
	"net/http"
	"time"

	change "github.com/iti/pbconf/lib/pbchange"
	database "github.com/iti/pbconf/lib/pbdatabase"
	global "github.com/iti/pbconf/lib/pbglobal"
	internode "github.com/iti/pbconf/lib/pbinternode"
	logging "github.com/iti/pbconf/lib/pblogger"
	policy "github.com/iti/pbconf/lib/pbpolicy"
)

type Heartbeat struct {
	log      logging.Logger
	db       database.AppDatabase
	commitId string
}

var nodeComm *internode.InterNode

func NewHeartbeatComm(d database.AppDatabase, loglevel string) *Heartbeat {
	l, _ := logging.GetLogger("Heartbeat")
	logging.SetLevel(loglevel, "Heartbeat")
	nodeComm = internode.NewInterNodeCommunicator(d)
	global.UpstreamConnectedStatus = false
	return &Heartbeat{log: l, db: d}
}

func (hb *Heartbeat) Start(doneChan chan bool) {
	go hb.startReportingTicker(doneChan)
	go hb.startPollingTicker(doneChan)
}

func (hb *Heartbeat) startReportingTicker(doneChan chan bool) {
	node, err := nodeComm.GetRootNode()
	if err != nil {
		hb.log.Warning("Report Ticker not started: Could not recover root node to start reporting ticker")
		return
	}
	duration, err := hb.db.GetDurationForTimer(node.Id, "ReportTimer")
	if err != nil {
		hb.log.Warning("Report Ticker not started: Could not get report timer duration from db to report to master periodically, Error: %s", err.Error())
		return
	}

	nBeatMultiplier := hb.db.GetTimeoutMultiplier(node.Id)
	timeoutDuration := time.Duration(float64(nBeatMultiplier)*duration.Seconds()) * time.Second

	reportTicker := time.NewTicker(duration)
	reportChan := reportTicker.C
	for {
		select {
		case <-reportChan:
			hb.onReportTick(timeoutDuration)
		case <-doneChan:
			reportTicker.Stop()
			return
		}
	}
}

func (hb *Heartbeat) startPollingTicker(doneChan chan bool) {
	node, err := nodeComm.GetRootNode()
	if err != nil {
		hb.log.Warning("Polling Ticker not started: Could not recover node to start polling ticker")
		return
	}

	pollDuration, err := hb.db.GetDurationForTimer(node.Id, "PollTimer")
	if err != nil {
		hb.log.Warning("Polling Ticker not started: Could not get poll timer duration from db to check on child nodes, Error: %s", err.Error())
		return
	}

	nBeatMultiplier := hb.db.GetTimeoutMultiplier(node.Id)
	timeoutDuration := time.Duration(float64(nBeatMultiplier)*pollDuration.Seconds()) * time.Second
	hb.log.Debug("timeout value for %v", timeoutDuration)

	pollTicker := time.NewTicker(pollDuration)
	pollChan := pollTicker.C

	for {
		select {
		case <-pollChan:
			hb.onPollTick(timeoutDuration)
		case <-doneChan:
			pollTicker.Stop()
			return
		}
	}
}

// onReportTick function reports to the upstream node when this function is invoked on report timer tick.
// The node keeps track of the last successful report time. Every time we successfully report upstream, we
// check the time between the last and current report time. If this time is greater than timeout, we assume
// it is a restored connection.
func (hb *Heartbeat) onReportTick(timeout time.Duration) {
	node, err := nodeComm.GetRootNode()
	if err != nil {
		hb.log.Warning("onReportTick::Could not recover node details from the database.")
		return
	}
	upstreamNodeIP, err := nodeComm.GetUpstreamNodeIP()
	if err != nil {
		hb.log.Info("onReportTick::Could not recover node IP address from the database.")
		return
	}
	if upstreamNodeIP == nil {
		return
	}
	//At this point an upstream node exists, we can try connecting to it
	var lastReportedTime time.Time
	reportCfgItem := database.PbNodeConfigItem{NodeId: node.Id, ConfigItem: database.ConfigItem{Key: "Last Reported Timestamp"}}
	reportCfgItemExists, err := reportCfgItem.Exists(hb.db)
	if err != nil {
		hb.log.Debug("Error trying to determine if we have reported to upstream node ever.Error:%s", err.Error())
	}
	if reportCfgItemExists {
		if err = reportCfgItem.Get(hb.db); err != nil {
			hb.log.Debug("Error trying to access last reported time: %s", err.Error())
			return
		}
		lastReportedTime, err = time.Parse(time.RFC1123, reportCfgItem.ConfigItem.Value)
		if err != nil {
			hb.log.Debug("Error parsing the time since last reporting time: %s", err.Error())
			return
		}
	}
	//Now try connecting to the upstream node
	destUrl := "https://" + *upstreamNodeIP + "/policy/default/" + node.Name
	req, err := http.NewRequest("HEAD", destUrl, nil)
	hb.log.Debug("onReportTick::Sending heartbeat to upstream node")
	resp, err := global.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		hb.log.Info("onReportTick::Connection to upstream node is lost")
		hb.log.Debug("onReportTick::Upstream node did not respond: %s", err.Error())
		if reportCfgItemExists && time.Since(lastReportedTime) > timeout {
			global.UpstreamConnectedStatus = false
		}
		return
	}
	//upstream node responded with ok or something, check it
	if resp.StatusCode == http.StatusOK {
		if !reportCfgItemExists { //first time we are ever reporting to an upstream node, add all devices
			hb.log.Debug("Syncing to upstream node for the first time!")
			err = hb.syncUpstreamNode(node.Id)
			internode.Qmutex.Lock()
			global.Queue.Init() //TO AVOID BUG WITH ORDER OF NODES STARTING DURING TESTING, ALSO FAIL SAFE
			internode.Qmutex.Unlock()
			if err != nil {
				return
			}
		} else {
			if time.Since(lastReportedTime) > timeout {
				hb.log.Info("Connection to upstream node restored. Resyncing to upstream now.")
				hb.resyncUpstreamNode()
			}
		}
		//update the last reported time
		reportCfgItem.ConfigItem.Value = time.Now().Format(time.RFC1123)
		reportCfgItem.CreateOrUpdate(hb.db)
		global.UpstreamConnectedStatus = true
		//check on policy status to see if we need to get updated policy (any policy change will trigger getting all policies)
		commitId := resp.Header.Get("X-Pbconf-Policy-LastCommitId")
		if commitId != hb.commitId {
			err := hb.updatePolicies()
			if err != nil {
			} else {
				hb.commitId = commitId
			}
		} //if commitId != hb.commitId
	} else {
		//Ccould not report to master node, log it and move on
		hb.log.Info("onReportTick::Could not successfully report to upstream node because of issues other than connection issues.")
		hb.log.Debug("Got response other than http.StatusOK from master node, got status:%s", resp.StatusCode)
	}
}

// syncUpstreamNode is called ONCE. The first time we connect to an upstream node. At this point,
// we try and update the information stored in the upstream node to include the port number at which
// this node can be contacted at, and send all the devices that this node has in its hierarachy.
func (hb *Heartbeat) syncUpstreamNode(rootnodeId int64) error {
	//send the port to the upstream node
	err := nodeComm.UpdateIPAddressUpstream()
	if err != nil {
		return err
	}
	upstreamNodeIPInfo, err := nodeComm.GetUpstreamNodeIP()
	if err != nil {
		return err
	}

	deviceList, err := hb.db.GetDevices()
	if err != nil {
		return err
	}
	for _, device := range deviceList {
		metaDevice := database.PbDevice{Id: device.Id}
		if err = metaDevice.Get(hb.db); err != nil {
			continue
		}
		nodeComm.AddDeviceUpstream(metaDevice)
	}
	if len(deviceList) > 0 && hb.db.GetPropogationStatus(rootnodeId) {
		upInfo := internode.UpstreamTransInfo{
			IpInfo:        *upstreamNodeIPInfo,
			TransactionId: "",
		}
		engine, err := change.GetCMEngine(nil)
		if err != nil {
			return err
		}
		engine.Push(change.DEVICE, upInfo, global.RootNode)
	}
	return nil
}

// resyncUpstreamNode is called every time that the connection is detected to have been restored after a
// loss of connection.
func (hb *Heartbeat) resyncUpstreamNode() error {
	internode.Qmutex.Lock()
	element := global.Queue.Front()
	internode.Qmutex.Unlock()
	var err error
	for {
		if element == nil {
			break
		}
		x := *element
		action, ok := x.Value.(internode.Qelement)
		put_on_q := false
		if ok {
			switch action.Fn_name {
			case "AddDeviceUpstream":
				put_on_q, err = nodeComm.AddDeviceUpstream(*action.Device)
			case "DeleteDeviceUpstream":
				put_on_q, err = nodeComm.DeleteDeviceUpstream(*action.Device)
			case "UpdateDeviceUpstream":
				put_on_q, err = nodeComm.UpdateDeviceUpstream(*action.Device, action.Param)
			case "DeleteDeviceConfigItemUpstream":
				put_on_q, err = nodeComm.DeleteDeviceConfigItemUpstream(*action.Device, action.Param)
			case "UpdateDeviceConfigUpstream":
				put_on_q, err = nodeComm.UpdateDeviceConfigUpstream(*action.Device, action.Cfg)
			case "UpdateDeviceMetaUpstream":
				put_on_q, err = nodeComm.UpdateDeviceMetaUpstream(*action.Device, action.Param, action.Param2)
			}
			if err != nil {
				//There's nothig to correct this condition. Out of sync for reasons not related to node communication
			}
		}
		if !put_on_q {
			internode.Qmutex.Lock()
			global.Queue.Remove(element)
			internode.Qmutex.Unlock()
		} else { //if we were unsuccessful at sending upstream because comms failed again, get out, things will remain on the queue
			break
		}
		internode.Qmutex.Lock()
		element = global.Queue.Front() //get next element
		internode.Qmutex.Unlock()
	}
	//at the end, we have successfully
	return nil
}

func (hb *Heartbeat) updatePolicies() error {
	upstreamPolicyList, err := nodeComm.GetUpstreamNodePolicies()
	if err != nil {
		hb.log.Debug("Error getting policies from upstream node, Error:%s", err)
		return err
	}

	engine, err := change.GetCMEngine(nil)
	if err != nil {
		return err
	}

	policyList, err := engine.ListObjects(change.POLICY)
	if err != nil {
		_, ok := err.(change.CMNoRepoError)
		if !ok {
			hb.log.Debug(err.Error())
			return err
		}
	}

	currentPolicySet := make(map[string]bool, len(policyList))
	for _, p := range policyList {
		currentPolicySet[p] = false
	}
	//loop through upstream policies and add or update contents
	for _, policy := range upstreamPolicyList {
		_, err = engine.VersionObject(&policy, policy.Log.Message)
		if err != nil {
		}
		_, ok := currentPolicySet[policy.Content.Object]
		if ok {
			currentPolicySet[policy.Content.Object] = true
		}
	}
	//now delete any policies that dont exist upstream anymore
	for pol, update_flag := range currentPolicySet {
		if update_flag == false {
			err = engine.RemoveObject(change.POLICY, pol, &change.CMAuthor{Name: "Policy engine"})
			if err != nil {
				hb.log.Debug("In heartbeat, error while trying to remove policy, Error:%s", err.Error())
				return err
			}
		}
	}
	status, explanation, err := policy.ValidateAgainstOntology(hb.log, nil)
	if err != nil {
		hb.log.Debug("Validating policy with ontology, got Error:%s from ontology engine", err.Error())
		return err
	}
	if status == false {
		return errors.New(fmt.Sprintf("Upstream node ontology is not valid, explanation = %s", explanation))
	}
	policy.LogOntologyInconsistencies(hb.log)

	return nil
}

// onPollTick function checks all the child nodes have checked in within the timeout
func (hb *Heartbeat) onPollTick(timeout time.Duration) {
	rootnode, err := nodeComm.GetRootNode()
	if err != nil {
		hb.log.Warning("onPollTick::Could not recover node details from the database")
		return
	}

	nodes, err := rootnode.GetChildNodes(hb.db)
	if err != nil {
		hb.log.Debug("onPollTick::Did not find any nodes to poll.")
		return
	}

	for _, childNode := range nodes {
		go func(chNode database.PbNode) {
			chCheckinItem := database.PbNodeConfigItem{NodeId: chNode.Id, ConfigItem: database.ConfigItem{Key: "Checkin Timestamp"}}
			exists, err := chCheckinItem.Exists(hb.db)
			if err != nil {
				hb.log.Info("onPollTick::Error checking existence of Checkin timestamp")
				return
			}
			if !exists {
				hb.log.Debug("onPollTick::Checkin timestamp does not exist yet")
				return
			}
			//timestamp exists, now get it
			err = chCheckinItem.Get(hb.db)
			if err != nil {
				hb.log.Info("onPollTick::Could not get Checkin timestamp Error: %s", err.Error())
				return
			}
			checkinTime, err := time.Parse(time.RFC1123, chCheckinItem.ConfigItem.Value)
			if err != nil {
				hb.log.Debug("onPollTick::time.Parse error :%s", err.Error())
				return
			}

			hb.log.Info("Node %s checkinTimeDuration = %v, timeout after %v", chNode.Name, time.Since(checkinTime), timeout)
			if time.Since(checkinTime) > timeout {
				// RAISE ALARM THAT a downstream node has gone offline
				hb.log.Error(fmt.Sprintf("Node %s has not checked in since %s. %s should now operate as master", chNode.Name, checkinTime.String(), chNode.Name))
			}
		}(childNode)
	}
}
