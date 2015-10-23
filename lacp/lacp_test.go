// test_lacp
package lacp_test

import (
	"fmt"
	"l2/lacp"
	"testing"
	"time"
)

func TestLaAggPortCreate(t *testing.T) {

	p := lacp.NewLaAggPort(1, "eth1.1")

	p.Start(false)

	fmt.Println("TEST: Port Enabled and Lacp Enabled -> RX Machine")
	p.RxMachineFsm.RxmEvents <- lacp.LacpRxmEventPortEnabledAndLacpEnabled

	time.Sleep(time.Second * 30)

}
