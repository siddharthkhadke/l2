Copyright [2016] [SnapRoute Inc]

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

	 Unless required by applicable law or agreed to in writing, software
	 distributed under the License is distributed on an "AS IS" BASIS,
	 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	 See the License for the specific language governing permissions and
	 limitations under the License.
// logger.go
package stp

import (
	"fmt"
	"strings"
)

func StpLogger(t string, msg string) {

	switch t {
	case "INFO":
		gLogger.Info(msg)
	case "DEBUG":
		gLogger.Debug(msg)
	case "ERROR":
		gLogger.Err(msg)
	case "WARNING":
		gLogger.Warning(msg)
	}
}

func StpLoggerInfo(msg string) {
	StpLogger("INFO", msg)
}

func StpMachineLogger(t string, m string, p int32, b int32, msg string) {
	StpLogger(t, strings.Join([]string{m, fmt.Sprintf("port %d", p), fmt.Sprintf("brg %d", b), msg}, ":"))
}
