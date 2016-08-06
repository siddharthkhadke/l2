//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//	 Unless required by applicable law or agreed to in writing, software
//	 distributed under the License is distributed on an "AS IS" BASIS,
//	 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	 See the License for the specific language governing permissions and
//	 limitations under the License.
//
// _______  __       __________   ___      _______.____    __    ____  __  .___________.  ______  __    __
// |   ____||  |     |   ____\  \ /  /     /       |\   \  /  \  /   / |  | |           | /      ||  |  |  |
// |  |__   |  |     |  |__   \  V  /     |   (----` \   \/    \/   /  |  | `---|  |----`|  ,----'|  |__|  |
// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |     |  |     |   __   |
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |     |  `----.|  |  |  |
// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__|
//
// config_test.go
package drcp

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"github.com/google/gopacket/layers"
	"l2/lacp/protocol/lacp"
	"l2/lacp/protocol/utils"
	"math"
	"net"
	"strings"
	"testing"
	"time"
	"utils/fsm"
	"utils/logging"
)

//const ipplink1 int32 = 3
//const aggport1 int32 = 1
//const aggport2 int32 = 2

//const ipplink2 int32 = 4
//const aggport3 int32 = 5
//const aggport4 int32 = 6

//type MyTestMock struct {
//	asicdmock.MockAsicdClientMgr
//}

func OnlyForRxMachineTestSetup() {
	logger, _ := logging.NewLogger("lacpd", "TEST", false)
	utils.SetLaLogger(logger)
	utils.SetAsicDPlugin(&MyTestMock{})
}

func OnlyForRxMachineTestTeardown() {

	utils.SetLaLogger(nil)
	utils.DeleteAllAsicDPlugins()
}

func OnlyForRxMachineCreateValidDRCPPacket() *layers.DRCP {

	phash := md5.New()
	for i := 0; i < MAX_CONVERSATION_IDS; i++ {

		if i != 100 {
			buf := new(bytes.Buffer)
			// network byte order
			binary.Write(buf, binary.BigEndian, []uint16{0x0000})
			phash.Write(buf.Bytes())
		} else {
			buf := new(bytes.Buffer)
			// network byte order
			binary.Write(buf, binary.BigEndian, []uint16{uint16(aggport3), uint16(aggport4)})
			phash.Write(buf.Bytes())
		}
	}

	ghash := md5.New()
	for i := float64(0); i < MAX_CONVERSATION_IDS; i++ {

		// even vlans owned by local system
		if math.Mod(i, 2) == 0 {
			buf := new(bytes.Buffer)
			// network byte order
			binary.Write(buf, binary.BigEndian, []uint8{1, 2})
			ghash.Write(buf.Bytes())
		} else {
			buf := new(bytes.Buffer)
			// network byte order
			binary.Write(buf, binary.BigEndian, []uint8{2, 1})
			ghash.Write(buf.Bytes())
		}
	}

	portdigest := phash.Sum(nil)
	gatewaydigest := ghash.Sum(nil)

	drcp := &layers.DRCP{
		PortalInfo: layers.DRCPPortalInfoTlv{
			TlvTypeLength:  layers.DRCPTLVTypePortalInfo | layers.DRCPTLVPortalInfoLength,
			AggPriority:    128,
			AggId:          [6]uint8{0x00, 0x00, 0x00, 0x00, 0x00, 0x64},
			PortalPriority: 128,
			PortalAddr:     [6]uint8{0x00, 0x00, 0xDE, 0xAD, 0xBE, 0xEF},
		},
		PortalConfigInfo: layers.DRCPPortalConfigurationInfoTlv{
			TlvTypeLength:    layers.DRCPTLVTypePortalConfigInfo | layers.DRCPTLVPortalConfigurationInfoLength,
			TopologyState:    layers.DRCPTopologyState(0x6),
			OperAggKey:       200,
			PortAlgorithm:    [4]uint8{0x00, 0x80, 0xC2, 0x01}, // C-VID
			GatewayAlgorithm: [4]uint8{0x00, 0x80, 0xC2, 0x01},
			PortDigest: [16]uint8{
				portdigest[0], portdigest[1], portdigest[2], portdigest[3],
				portdigest[4], portdigest[5], portdigest[6], portdigest[7],
				portdigest[8], portdigest[9], portdigest[10], portdigest[11],
				portdigest[12], portdigest[13], portdigest[14], portdigest[15],
			},
			GatewayDigest: [16]uint8{
				gatewaydigest[0], gatewaydigest[1], gatewaydigest[2], gatewaydigest[3],
				gatewaydigest[4], gatewaydigest[5], gatewaydigest[6], gatewaydigest[7],
				gatewaydigest[8], gatewaydigest[9], gatewaydigest[10], gatewaydigest[11],
				gatewaydigest[12], gatewaydigest[13], gatewaydigest[14], gatewaydigest[15],
			},
		},
		State: layers.DRCPStateTlv{
			TlvTypeLength: layers.DRCPTLVTypeDRCPState | layers.DRCPTLVStateLength,
			State:         layers.DRCPState(1 << layers.DRCPStateHomeGatewayBit),
		},
		HomePortsInfo: layers.DRCPHomePortsInfoTlv{
			TlvTypeLength:     layers.DRCPTLVTypeHomePortsInfo | layers.DRCPTlvTypeLength(8),
			AdminAggKey:       100,
			OperPartnerAggKey: 100,
			ActiveHomePorts:   []uint32{uint32(aggport3), uint32(aggport4)},
		},
		//NeighborPortsInfo:                  DRCPNeighborPortsInfoTlv{},
		HomeGatewayVector: layers.DRCPHomeGatewayVectorTlv{
			TlvTypeLength: layers.DRCPTLVTypeHomeGatewayVector | layers.DRCPTLVHomeGatewayVectorLength_2,
			Sequence:      1,
			Vector:        make([]uint8, 4096),
		},
		//NeighborGatewayVector: DRCPNeighborGatewayVectorTlv{
		//	TlvTypeLength: layers.DRCPTLVTypeNeighborGatewayVector | layers.DRCPTLVNeighborGatewayVectorLength,
		//	Sequence:      1,
		//},
		//TwoPortalGatewayConversationVector: DRCP2PGatewayConversationVectorTlv{},
		//TwoPortalPortConversationVector:    DRCP2PPortConversationVectorTlv{},
		NetworkIPLMethod: layers.DRCPNetworkIPLSharingMethodTlv{
			TlvTypeLength: layers.DRCPTLVNetworkIPLSharingMethod | layers.DRCPTLVNetworkIPLSharingMethodLength,
			Method:        [4]uint8{0x00, 0x80, 0xC2, 0x1},
		},
		//NetworkIPLEncapsulation:            DRCPNetworkIPLSharingEncapsulationTlv{},
	}
	return drcp
}

func OnlyForRxMachineTestSetupCreateAggGroup(aggId uint32) *lacp.LaAggregator {
	a1conf := &lacp.LaAggConfig{
		Mac: [6]uint8{0x00, 0x00, 0x01, 0x01, 0x01, 0x01},
		Id:  int(aggId),
		Key: uint16(aggId),
		Lacp: lacp.LacpConfigInfo{Interval: lacp.LacpFastPeriodicTime,
			Mode:           lacp.LacpModeActive,
			SystemIdMac:    "00:00:00:00:00:64",
			SystemPriority: 128},
	}
	lacp.CreateLaAgg(a1conf)

	p1conf := &lacp.LaAggPortConfig{
		Id:      uint16(aggport1),
		Prio:    128,
		Key:     uint16(aggId),
		AggId:   int(aggId),
		Enable:  true,
		Mode:    lacp.LacpModeActive,
		Timeout: lacp.LacpShortTimeoutTime,
		Properties: lacp.PortProperties{
			Mac:    net.HardwareAddr{0x00, byte(aggport1), 0xDE, 0xAD, 0xBE, 0xEF},
			Speed:  1000000000,
			Duplex: lacp.LacpPortDuplexFull,
			Mtu:    1500,
		},
		IntfId:   utils.PortConfigMap[aggport1].Name,
		TraceEna: false,
	}

	lacp.CreateLaAggPort(p1conf)
	lacp.AddLaAggPortToAgg(a1conf.Key, p1conf.Id)

	var a *lacp.LaAggregator
	if lacp.LaFindAggById(a1conf.Id, &a) {
		return a
	}
	return nil
}

func RxMachineTestSetup() {
	OnlyForRxMachineTestSetup()
	utils.PortConfigMap[ipplink1] = utils.PortConfig{Name: "SIMeth1.3",
		HardwareAddr: net.HardwareAddr{0x00, 0x33, 0x11, 0x22, 0x22, 0x33},
	}
	utils.PortConfigMap[aggport1] = utils.PortConfig{Name: "SIMeth1.1",
		HardwareAddr: net.HardwareAddr{0x00, 0x11, 0x11, 0x22, 0x22, 0x33},
	}
	utils.PortConfigMap[aggport2] = utils.PortConfig{Name: "SIMeth1.2",
		HardwareAddr: net.HardwareAddr{0x00, 0x22, 0x11, 0x22, 0x22, 0x33},
	}
	utils.PortConfigMap[ipplink2] = utils.PortConfig{Name: "SIMeth0.3",
		HardwareAddr: net.HardwareAddr{0x00, 0x44, 0x11, 0x22, 0x22, 0x33},
	}
	utils.PortConfigMap[aggport3] = utils.PortConfig{Name: "SIMeth0.1",
		HardwareAddr: net.HardwareAddr{0x00, 0x55, 0x11, 0x22, 0x22, 0x33},
	}
	utils.PortConfigMap[aggport4] = utils.PortConfig{Name: "SIMeth0.2",
		HardwareAddr: net.HardwareAddr{0x00, 0x66, 0x11, 0x22, 0x22, 0x33},
	}
}
func RxMachineTestTeardwon() {

	OnlyForRxMachineTestTeardown()
	delete(utils.PortConfigMap, ipplink1)
	delete(utils.PortConfigMap, aggport1)
	delete(utils.PortConfigMap, aggport2)
	delete(utils.PortConfigMap, ipplink2)
	delete(utils.PortConfigMap, aggport3)
	delete(utils.PortConfigMap, aggport4)
}

func TestRxMachineRxValidDRCPDUNeighborPkt(t *testing.T) {

	RxMachineTestSetup()
	a := OnlyForRxMachineTestSetupCreateAggGroup(200)

	cfg := &DistrubtedRelayConfig{
		DrniName:                          "DR-1",
		DrniPortalAddress:                 "00:00:DE:AD:BE:EF",
		DrniPortalPriority:                128,
		DrniThreePortalSystem:             false,
		DrniPortalSystemNumber:            1,
		DrniIntraPortalLinkList:           [3]uint32{uint32(ipplink1)},
		DrniAggregator:                    uint32(a.AggId),
		DrniGatewayAlgorithm:              "00:80:C2:01",
		DrniNeighborAdminGatewayAlgorithm: "00:80:C2:01",
		DrniNeighborAdminPortAlgorithm:    "00:80:C2:01",
		DrniNeighborAdminDRCPState:        "00000000",
		DrniEncapMethod:                   "00:80:C2:01",
		DrniPortConversationControl:       false,
		DrniIntraPortalPortProtocolDA:     "01:80:C2:00:00:03", // only supported value that we are going to support
	}
	// map vlan 100 to this system
	// in real system this should be filled in by vlan membership
	cfg.DrniConvAdminGateway[100][0] = cfg.DrniPortalSystemNumber

	err := DistrubtedRelayConfigParamCheck(cfg)
	if err != nil {
		t.Error("Parameter check failed for what was expected to be a valid config", err)
	}
	// just create instance not starting any state machines
	dr := NewDistributedRelay(cfg)
	dr.a = a

	// create packet which will be sent for this test
	drcp := OnlyForRxMachineCreateValidDRCPPacket()

	dr.DRFHomeConversationPortListDigest = drcp.PortalConfigInfo.PortDigest
	dr.DRFHomeConversationGatewayListDigest = drcp.PortalConfigInfo.GatewayDigest

	// lets get the IPP
	ipp := dr.Ipplinks[0]

	// rx machine sends event to each of these machines according to figure 9-22
	DrcpAMachineFSMBuild(dr)
	DrcpGMachineFSMBuild(dr)
	DrcpPsMachineFSMBuild(dr)
	DrcpTxMachineFSMBuild(ipp)
	DrcpPtxMachineFSMBuild(ipp)

	// start RX MAIN
	ipp.DrcpRxMachineMain()
	// enable because aggregator was attached above
	ipp.DRCPEnabled = true

	responseChan := make(chan string)

	// Psm is expected to be in update state
	// initialize will set the dr default values
	// which will be matched against received packets
	dr.PsMachineFsm.DrcpPsMachinePortalSystemInitialize(*dr.PsMachineFsm.Machine, nil)

	ipp.RxMachineFsm.RxmEvents <- utils.MachineEvent{
		E:            RxmEventBegin,
		Src:          "RX MACHINE TEST",
		ResponseChan: responseChan,
	}

	<-responseChan

	ipp.RxMachineFsm.RxmPktRxEvent <- RxDrcpPdu{
		pdu:          drcp,
		src:          "RX MACHINE TEST",
		responseChan: responseChan,
	}

	<-responseChan

	// Neighbor Admin values not correct, thus should discard as
	// neighbor info is not known yet
	if ipp.RxMachineFsm.Machine.Curr.CurrentState() != RxmStateCurrent {
		t.Error("ERROR Rx Machine is not in expected state from first received PDU actual:", RxmStateStrMap[ipp.RxMachineFsm.Machine.Curr.CurrentState()])
	}
	// lets check some settings on the ipp
	if !ipp.DRFNeighborOperDRCPState.GetState(layers.DRCPStateIPPActivity) {
		t.Error("ERROR Neighbor_Oper_DRCP_State IPP_Activity was not set when it should be set", ipp.DRFNeighborOperDRCPState)
	}
	if !ipp.DRFNeighborOperDRCPState.GetState(layers.DRCPStateDRCPTimeout) {
		t.Error("ERROR Neighbor_Oper_DRCP_State DRCP Timeout was set to LONG when it should be SHORT", ipp.DRFNeighborOperDRCPState)
	}
	if ipp.DifferPortal {
		t.Error("ERROR packet portal info should agree with local since they are provisioned the same")
	}
	if ipp.DifferPortalReason != "" {
		t.Error("ERROR Portal Difference Detected", ipp.DifferPortalReason)
	}

	eventReceived := false
	go func(evrx *bool) {
		for i := 0; i < 10 && !*evrx; i++ {
			time.Sleep(time.Second * 1)
		}
		if !eventReceived {
			ipp.TxMachineFsm.TxmEvents <- utils.MachineEvent{
				E:   fsm.Event(0),
				Src: "RX MACHINE: FORCE TEST FAIL",
			}
		}
	}(&eventReceived)

	evt := <-ipp.TxMachineFsm.TxmEvents
	if evt.E != TxmEventNtt {
		t.Error("ERROR Invalid event received", ipp.DifferPortalReason)
	}

	lacp.DeleteLaAgg(a.AggId)
	RxMachineTestTeardwon()
}

func TestRxMachineRxPktDRCPDUNeighborPortalInfoDifferAggregatorPriority(t *testing.T) {

	RxMachineTestSetup()
	a := OnlyForRxMachineTestSetupCreateAggGroup(100)

	cfg := &DistrubtedRelayConfig{
		DrniName:                          "DR-1",
		DrniPortalAddress:                 "00:00:DE:AD:BE:EF",
		DrniPortalPriority:                128,
		DrniThreePortalSystem:             false,
		DrniPortalSystemNumber:            1,
		DrniIntraPortalLinkList:           [3]uint32{uint32(ipplink1)},
		DrniAggregator:                    uint32(a.AggId),
		DrniGatewayAlgorithm:              "00:80:C2:01",
		DrniNeighborAdminGatewayAlgorithm: "00:80:C2:01",
		DrniNeighborAdminPortAlgorithm:    "00:80:C2:01",
		DrniNeighborAdminDRCPState:        "00000000",
		DrniEncapMethod:                   "00:80:C2:01",
		DrniPortConversationControl:       false,
		DrniIntraPortalPortProtocolDA:     "01:80:C2:00:00:03", // only supported value that we are going to support
	}
	// map vlan 100 to this system
	// in real system this should be filled in by vlan membership
	cfg.DrniConvAdminGateway[100][0] = cfg.DrniPortalSystemNumber

	err := DistrubtedRelayConfigParamCheck(cfg)
	if err != nil {
		t.Error("Parameter check failed for what was expected to be a valid config", err)
	}
	// just create instance not starting any state machines
	dr := NewDistributedRelay(cfg)
	dr.a = a

	// lets get the IPP
	ipp := dr.Ipplinks[0]

	// rx machine sends event to each of these machines according to figure 9-22
	DrcpAMachineFSMBuild(dr)
	DrcpGMachineFSMBuild(dr)
	DrcpPsMachineFSMBuild(dr)
	DrcpTxMachineFSMBuild(ipp)
	DrcpPtxMachineFSMBuild(ipp)

	// start RX MAIN
	ipp.DrcpRxMachineMain()
	// enable because aggregator was attached above
	ipp.DRCPEnabled = true

	responseChan := make(chan string)

	// Psm is expected to be in update state
	// initialize will set the dr default values
	// which will be matched against received packets
	dr.PsMachineFsm.DrcpPsMachinePortalSystemInitialize(*dr.PsMachineFsm.Machine, nil)

	ipp.RxMachineFsm.RxmEvents <- utils.MachineEvent{
		E:            RxmEventBegin,
		Src:          "RX MACHINE TEST",
		ResponseChan: responseChan,
	}

	<-responseChan

	// create packet
	drcp := OnlyForRxMachineCreateValidDRCPPacket()

	drcp.PortalInfo.AggPriority = 256

	ipp.RxMachineFsm.RxmPktRxEvent <- RxDrcpPdu{
		pdu:          drcp,
		src:          "RX MACHINE TEST",
		responseChan: responseChan,
	}

	<-responseChan

	// Neighbor Admin values not correct, thus should discard as
	// neighbor info is not known yet
	if ipp.RxMachineFsm.Machine.Curr.CurrentState() != RxmStateDiscard {
		t.Error("ERROR Rx Machine is not in expected state from first received PDU actual:", RxmStateStrMap[ipp.RxMachineFsm.Machine.Curr.CurrentState()])
	}
	// lets check some settings on the ipp
	if ipp.DRFNeighborOperDRCPState.GetState(layers.DRCPStateIPPActivity) {
		t.Error("ERROR Neighbor_Oper_DRCP_State IPP_Activity was set when it should be cleared", ipp.DRFNeighborOperDRCPState)
	}
	if !ipp.DRFNeighborOperDRCPState.GetState(layers.DRCPStateDRCPTimeout) {
		t.Error("ERROR Neighbor_Oper_DRCP_State DRCP Timeout was set to LONG when it should be SHORT", ipp.DRFNeighborOperDRCPState)
	}
	if !ipp.DifferPortal {
		t.Error("ERROR packet portal info should not agree with local since they are provisioned differently")
	}
	if !strings.Contains(ipp.DifferPortalReason, "Neighbor Aggregator Priority") {
		t.Error("ERROR Portal Difference Detected", ipp.DifferPortalReason)
	}

	lacp.DeleteLaAgg(a.AggId)
	RxMachineTestTeardwon()
}

func TestRxMachineRxPktDRCPDUNeighborPortalInfoDifferAggregatorAddr(t *testing.T) {

	RxMachineTestSetup()
	a := OnlyForRxMachineTestSetupCreateAggGroup(100)

	cfg := &DistrubtedRelayConfig{
		DrniName:                          "DR-1",
		DrniPortalAddress:                 "00:00:DE:AD:BE:EF",
		DrniPortalPriority:                128,
		DrniThreePortalSystem:             false,
		DrniPortalSystemNumber:            1,
		DrniIntraPortalLinkList:           [3]uint32{uint32(ipplink1)},
		DrniAggregator:                    uint32(a.AggId),
		DrniGatewayAlgorithm:              "00:80:C2:01",
		DrniNeighborAdminGatewayAlgorithm: "00:80:C2:01",
		DrniNeighborAdminPortAlgorithm:    "00:80:C2:01",
		DrniNeighborAdminDRCPState:        "00000000",
		DrniEncapMethod:                   "00:80:C2:01",
		DrniPortConversationControl:       false,
		DrniIntraPortalPortProtocolDA:     "01:80:C2:00:00:03", // only supported value that we are going to support
	}
	// map vlan 100 to this system
	// in real system this should be filled in by vlan membership
	cfg.DrniConvAdminGateway[100][0] = cfg.DrniPortalSystemNumber

	err := DistrubtedRelayConfigParamCheck(cfg)
	if err != nil {
		t.Error("Parameter check failed for what was expected to be a valid config", err)
	}
	// just create instance not starting any state machines
	dr := NewDistributedRelay(cfg)
	dr.a = a

	// lets get the IPP
	ipp := dr.Ipplinks[0]

	// rx machine sends event to each of these machines according to figure 9-22
	DrcpAMachineFSMBuild(dr)
	DrcpGMachineFSMBuild(dr)
	DrcpPsMachineFSMBuild(dr)
	DrcpTxMachineFSMBuild(ipp)
	DrcpPtxMachineFSMBuild(ipp)

	// start RX MAIN
	ipp.DrcpRxMachineMain()
	// enable because aggregator was attached above
	ipp.DRCPEnabled = true

	responseChan := make(chan string)

	// Psm is expected to be in update state
	// initialize will set the dr default values
	// which will be matched against received packets
	dr.PsMachineFsm.DrcpPsMachinePortalSystemInitialize(*dr.PsMachineFsm.Machine, nil)

	ipp.RxMachineFsm.RxmEvents <- utils.MachineEvent{
		E:            RxmEventBegin,
		Src:          "RX MACHINE TEST",
		ResponseChan: responseChan,
	}

	<-responseChan

	// create packet
	drcp := OnlyForRxMachineCreateValidDRCPPacket()

	drcp.PortalInfo.AggId = [6]uint8{0x00, 0x00, 0x00, 0x11, 0x00, 0x64}

	ipp.RxMachineFsm.RxmPktRxEvent <- RxDrcpPdu{
		pdu:          drcp,
		src:          "RX MACHINE TEST",
		responseChan: responseChan,
	}

	<-responseChan

	// Neighbor Admin values not correct, thus should discard as
	// neighbor info is not known yet
	if ipp.RxMachineFsm.Machine.Curr.CurrentState() != RxmStateDiscard {
		t.Error("ERROR Rx Machine is not in expected state from first received PDU actual:", RxmStateStrMap[ipp.RxMachineFsm.Machine.Curr.CurrentState()])
	}
	// lets check some settings on the ipp
	if ipp.DRFNeighborOperDRCPState.GetState(layers.DRCPStateIPPActivity) {
		t.Error("ERROR Neighbor_Oper_DRCP_State IPP_Activity was set when it should be cleared", ipp.DRFNeighborOperDRCPState)
	}
	if !ipp.DRFNeighborOperDRCPState.GetState(layers.DRCPStateDRCPTimeout) {
		t.Error("ERROR Neighbor_Oper_DRCP_State DRCP Timeout was set to LONG when it should be SHORT", ipp.DRFNeighborOperDRCPState)
	}
	if !ipp.DifferPortal {
		t.Error("ERROR packet portal info should not agree with local since they are provisioned differently")
	}
	if !strings.Contains(ipp.DifferPortalReason, "Neighbor Aggregator Id") {
		t.Error("ERROR Portal Difference Detected", ipp.DifferPortalReason)
	}

	lacp.DeleteLaAgg(a.AggId)
	RxMachineTestTeardwon()
}
