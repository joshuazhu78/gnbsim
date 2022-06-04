// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/factory"
	profctx "github.com/omec-project/gnbsim/profile/context"
	"github.com/omec-project/gnbsim/profile/util"
	"github.com/omec-project/gnbsim/simue"
	simuectx "github.com/omec-project/gnbsim/simue/context"
	"github.com/omec-project/gnbsim/profile/ransim"
)

//profile names
const (
	REGISTER                string = "register"
	PDU_SESS_EST            string = "pdusessest"
	DEREGISTER              string = "deregister"
	AN_RELEASE              string = "anrelease"
	UE_TRIGG_SERVICE_REQ    string = "uetriggservicereq"
	NW_TRIGG_UE_DEREG       string = "nwtriggeruedereg"
	UE_REQ_PDU_SESS_RELEASE string = "uereqpdusessrelease"
	NW_REQ_PDU_SESS_RELEASE string = "nwreqpdusessrelease"
	NW_REQ_PDU_SESS_MODIFY  string = "nwreqpdusessmodify"
)

func InitializeAllProfiles() {
	for _, profile := range factory.AppConfig.Configuration.Profiles {
		profile.Init()
	}
}

func ExecuteProfile(profile *profctx.Profile, summaryChan chan common.InterfaceMessage) {
	profile.Log.Infoln("executing profile:", profile.Name,
		", profile type:", profile.ProfileType)
	initEventMap(profile)
	initProcedureList(profile)

	if profile.PerUserTimeout == 0 {
		profile.PerUserTimeout = profctx.PER_USER_TIMEOUT
	}
	gnb, err := factory.AppConfig.Configuration.GetGNodeB(profile.GnbName)
	if err != nil {
		profile.Log.Errorln("GetGNodeB returned:", err)
	}

	imsi, err := strconv.Atoi(profile.StartImsi)
	if err != nil {
		profile.Log.Fatalln("invalid imsi value")
	}
	var wg sync.WaitGroup
	summary := &common.SummaryMessage{
		ProfileType: profile.ProfileType,
		ProfileName: profile.Name,
		ErrorList:   make([]error, 0, 10),
	}

	simUeChan := make(chan*simuectx.SimUe)
	go ransim.RunGetUEsCommand(simUeChan)
	// Currently executing profile for one IMSI at a time
	for count := 1; count <= profile.UeCount; count++ {
		imsiStr := "imsi-" + strconv.Itoa(imsi)
		simUe := simuectx.NewSimUe(imsiStr, gnb, profile)
		simUeChan <- simUe

		wg.Add(1)
		go func() {
			defer wg.Done()
			simue.Init(simUe)
		}()
		util.SendToSimUe(simUe, common.PROFILE_START_EVENT)

		timeout := time.Duration(profile.PerUserTimeout) * time.Second
		ticker := time.NewTicker(timeout)

		select {
		case <-ticker.C:
			err := fmt.Errorf("imsi:%v, profile timeout", imsiStr)
			profile.Log.Infoln("Result: FAIL,", err)
			summary.UeFailedCount++
			summary.ErrorList = append(summary.ErrorList, err)
			util.SendToSimUe(simUe, common.QUIT_EVENT)

		case msg := <-profile.ReadChan:
			switch msg.Event {
			case common.PROFILE_PASS_EVENT:
				profile.Log.Infoln("Result: PASS, imsi:", msg.Supi)
				summary.UePassedCount++
			case common.PROFILE_FAIL_EVENT:
				err := fmt.Errorf("imsi:%v, procedure:%v, error:%v", msg.Supi, msg.Proc, msg.Error)
				profile.Log.Infoln("Result: FAIL,", err)
				summary.UeFailedCount++
				summary.ErrorList = append(summary.ErrorList, err)
			}
		}
		ticker.Stop()
		time.Sleep(2 * time.Second)
		imsi++
	}

	summaryChan <- summary
	wg.Wait()
}

func initEventMap(profile *profctx.Profile) {
	switch profile.ProfileType {
	case REGISTER:
		profile.Events = map[common.EventType]common.EventType{
			common.REG_REQUEST_EVENT:     common.AUTH_REQUEST_EVENT,
			common.AUTH_REQUEST_EVENT:    common.AUTH_RESPONSE_EVENT,
			common.SEC_MOD_COMMAND_EVENT: common.SEC_MOD_COMPLETE_EVENT,
			common.REG_ACCEPT_EVENT:      common.REG_COMPLETE_EVENT,
			common.PROFILE_PASS_EVENT:    common.QUIT_EVENT,
		}
	case PDU_SESS_EST:
		profile.Events = map[common.EventType]common.EventType{
			common.REG_REQUEST_EVENT:          common.AUTH_REQUEST_EVENT,
			common.AUTH_REQUEST_EVENT:         common.AUTH_RESPONSE_EVENT,
			common.SEC_MOD_COMMAND_EVENT:      common.SEC_MOD_COMPLETE_EVENT,
			common.REG_ACCEPT_EVENT:           common.REG_COMPLETE_EVENT,
			common.PDU_SESS_EST_REQUEST_EVENT: common.PDU_SESS_EST_ACCEPT_EVENT,
			common.PDU_SESS_EST_ACCEPT_EVENT:  common.PDU_SESS_EST_ACCEPT_EVENT,
			common.PROFILE_PASS_EVENT:         common.QUIT_EVENT,
		}
	case UE_REQ_PDU_SESS_RELEASE:
		profile.Events = map[common.EventType]common.EventType{
			common.REG_REQUEST_EVENT:          common.AUTH_REQUEST_EVENT,
			common.AUTH_REQUEST_EVENT:         common.AUTH_RESPONSE_EVENT,
			common.SEC_MOD_COMMAND_EVENT:      common.SEC_MOD_COMPLETE_EVENT,
			common.REG_ACCEPT_EVENT:           common.REG_COMPLETE_EVENT,
			common.PDU_SESS_EST_REQUEST_EVENT: common.PDU_SESS_EST_ACCEPT_EVENT,
			common.PDU_SESS_EST_ACCEPT_EVENT:  common.PDU_SESS_EST_ACCEPT_EVENT,
			common.PDU_SESS_REL_REQUEST_EVENT: common.PDU_SESS_REL_COMMAND_EVENT,
			common.PDU_SESS_REL_COMMAND_EVENT: common.PDU_SESS_REL_COMPLETE_EVENT,
			common.PROFILE_PASS_EVENT:         common.QUIT_EVENT,
		}
	case DEREGISTER:
		profile.Events = map[common.EventType]common.EventType{
			common.REG_REQUEST_EVENT:           common.AUTH_REQUEST_EVENT,
			common.AUTH_REQUEST_EVENT:          common.AUTH_RESPONSE_EVENT,
			common.SEC_MOD_COMMAND_EVENT:       common.SEC_MOD_COMPLETE_EVENT,
			common.REG_ACCEPT_EVENT:            common.REG_COMPLETE_EVENT,
			common.PDU_SESS_EST_REQUEST_EVENT:  common.PDU_SESS_EST_ACCEPT_EVENT,
			common.PDU_SESS_EST_ACCEPT_EVENT:   common.PDU_SESS_EST_ACCEPT_EVENT,
			common.DEREG_REQUEST_UE_ORIG_EVENT: common.DEREG_ACCEPT_UE_ORIG_EVENT,
			common.PROFILE_PASS_EVENT:          common.QUIT_EVENT,
		}
	case AN_RELEASE:
		profile.Events = map[common.EventType]common.EventType{
			common.REG_REQUEST_EVENT:          common.AUTH_REQUEST_EVENT,
			common.AUTH_REQUEST_EVENT:         common.AUTH_RESPONSE_EVENT,
			common.SEC_MOD_COMMAND_EVENT:      common.SEC_MOD_COMPLETE_EVENT,
			common.REG_ACCEPT_EVENT:           common.REG_COMPLETE_EVENT,
			common.PDU_SESS_EST_REQUEST_EVENT: common.PDU_SESS_EST_ACCEPT_EVENT,
			common.PDU_SESS_EST_ACCEPT_EVENT:  common.PDU_SESS_EST_ACCEPT_EVENT,
			common.TRIGGER_AN_RELEASE_EVENT:   common.CONNECTION_RELEASE_REQUEST_EVENT,
			common.PROFILE_PASS_EVENT:         common.QUIT_EVENT,
		}
	case UE_TRIGG_SERVICE_REQ:
		profile.Events = map[common.EventType]common.EventType{
			common.REG_REQUEST_EVENT:          common.AUTH_REQUEST_EVENT,
			common.AUTH_REQUEST_EVENT:         common.AUTH_RESPONSE_EVENT,
			common.SEC_MOD_COMMAND_EVENT:      common.SEC_MOD_COMPLETE_EVENT,
			common.REG_ACCEPT_EVENT:           common.REG_COMPLETE_EVENT,
			common.PDU_SESS_EST_REQUEST_EVENT: common.PDU_SESS_EST_ACCEPT_EVENT,
			common.PDU_SESS_EST_ACCEPT_EVENT:  common.PDU_SESS_EST_ACCEPT_EVENT,
			common.SERVICE_REQUEST_EVENT:      common.SERVICE_ACCEPT_EVENT,
			common.TRIGGER_AN_RELEASE_EVENT:   common.CONNECTION_RELEASE_REQUEST_EVENT,
			common.PROFILE_PASS_EVENT:         common.QUIT_EVENT,
		}
	case NW_TRIGG_UE_DEREG:
		profile.Events = map[common.EventType]common.EventType{
			common.REG_REQUEST_EVENT:           common.AUTH_REQUEST_EVENT,
			common.AUTH_REQUEST_EVENT:          common.AUTH_RESPONSE_EVENT,
			common.SEC_MOD_COMMAND_EVENT:       common.SEC_MOD_COMPLETE_EVENT,
			common.REG_ACCEPT_EVENT:            common.REG_COMPLETE_EVENT,
			common.PDU_SESS_EST_REQUEST_EVENT:  common.PDU_SESS_EST_ACCEPT_EVENT,
			common.PDU_SESS_EST_ACCEPT_EVENT:   common.PDU_SESS_EST_ACCEPT_EVENT,
			common.DEREG_REQUEST_UE_TERM_EVENT: common.DEREG_ACCEPT_UE_TERM_EVENT,
			common.PROFILE_PASS_EVENT:          common.QUIT_EVENT,
		}
	case NW_REQ_PDU_SESS_RELEASE:
		profile.Events = map[common.EventType]common.EventType{
			common.REG_REQUEST_EVENT:          common.AUTH_REQUEST_EVENT,
			common.AUTH_REQUEST_EVENT:         common.AUTH_RESPONSE_EVENT,
			common.SEC_MOD_COMMAND_EVENT:      common.SEC_MOD_COMPLETE_EVENT,
			common.REG_ACCEPT_EVENT:           common.REG_COMPLETE_EVENT,
			common.PDU_SESS_EST_REQUEST_EVENT: common.PDU_SESS_EST_ACCEPT_EVENT,
			common.PDU_SESS_EST_ACCEPT_EVENT:  common.PDU_SESS_EST_ACCEPT_EVENT,
			common.PDU_SESS_REL_COMMAND_EVENT: common.PDU_SESS_REL_COMPLETE_EVENT,
			common.PROFILE_PASS_EVENT:         common.QUIT_EVENT,
		}
	case NW_REQ_PDU_SESS_MODIFY:
		profile.Events = map[common.EventType]common.EventType{
			common.REG_REQUEST_EVENT:          common.AUTH_REQUEST_EVENT,
			common.AUTH_REQUEST_EVENT:         common.AUTH_RESPONSE_EVENT,
			common.SEC_MOD_COMMAND_EVENT:      common.SEC_MOD_COMPLETE_EVENT,
			common.REG_ACCEPT_EVENT:           common.REG_COMPLETE_EVENT,
			common.PDU_SESS_EST_REQUEST_EVENT: common.PDU_SESS_EST_ACCEPT_EVENT,
			common.PDU_SESS_EST_ACCEPT_EVENT:  common.PDU_SESS_EST_ACCEPT_EVENT,
			common.PDU_SESS_MOD_REQUEST_EVENT: common.PDU_SESS_MOD_COMMAND_EVENT,
			common.PDU_SESS_MOD_COMMAND_EVENT: common.PDU_SESS_MOD_COMPLETE_EVENT,
			common.PROFILE_PASS_EVENT:         common.QUIT_EVENT,
		}
	}
}

func initProcedureList(profile *profctx.Profile) {
	switch profile.ProfileType {
	case REGISTER:
		profile.Procedures = []common.ProcedureType{common.REGISTRATION_PROCEDURE}
	case PDU_SESS_EST:
		profile.Procedures = []common.ProcedureType{
			common.REGISTRATION_PROCEDURE,
			common.PDU_SESSION_ESTABLISHMENT_PROCEDURE,
			common.USER_DATA_PKT_GENERATION_PROCEDURE,
		}
	case DEREGISTER:
		profile.Procedures = []common.ProcedureType{
			common.REGISTRATION_PROCEDURE,
			common.PDU_SESSION_ESTABLISHMENT_PROCEDURE,
			common.USER_DATA_PKT_GENERATION_PROCEDURE,
			common.UE_INITIATED_DEREGISTRATION_PROCEDURE,
		}
	case AN_RELEASE:
		profile.Procedures = []common.ProcedureType{
			common.REGISTRATION_PROCEDURE,
			common.PDU_SESSION_ESTABLISHMENT_PROCEDURE,
			common.USER_DATA_PKT_GENERATION_PROCEDURE,
			common.AN_RELEASE_PROCEDURE,
		}
	case UE_TRIGG_SERVICE_REQ:
		profile.Procedures = []common.ProcedureType{
			common.REGISTRATION_PROCEDURE,
			common.PDU_SESSION_ESTABLISHMENT_PROCEDURE,
			common.USER_DATA_PKT_GENERATION_PROCEDURE,
			common.AN_RELEASE_PROCEDURE,
			common.UE_TRIGGERED_SERVICE_REQUEST_PROCEDURE,
		}
	case NW_TRIGG_UE_DEREG:
		profile.Procedures = []common.ProcedureType{
			common.REGISTRATION_PROCEDURE,
			common.PDU_SESSION_ESTABLISHMENT_PROCEDURE,
			common.USER_DATA_PKT_GENERATION_PROCEDURE,
			common.NW_TRIGGERED_UE_DEREGISTRATION_PROCEDURE,
		}
	case UE_REQ_PDU_SESS_RELEASE:
		profile.Procedures = []common.ProcedureType{
			common.REGISTRATION_PROCEDURE,
			common.PDU_SESSION_ESTABLISHMENT_PROCEDURE,
			common.USER_DATA_PKT_GENERATION_PROCEDURE,
			common.UE_REQUESTED_PDU_SESSION_RELEASE_PROCEDURE,
		}
	case NW_REQ_PDU_SESS_RELEASE:
		profile.Procedures = []common.ProcedureType{
			common.REGISTRATION_PROCEDURE,
			common.PDU_SESSION_ESTABLISHMENT_PROCEDURE,
			common.USER_DATA_PKT_GENERATION_PROCEDURE,
			common.NW_REQUESTED_PDU_SESSION_RELEASE_PROCEDURE,
		}
	case NW_REQ_PDU_SESS_MODIFY:
		profile.Procedures = []common.ProcedureType{
			common.REGISTRATION_PROCEDURE,
			common.PDU_SESSION_ESTABLISHMENT_PROCEDURE,
			common.USER_DATA_PKT_GENERATION_PROCEDURE,
		}
	}
}
