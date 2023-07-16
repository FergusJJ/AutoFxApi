package handler

import (
	monitormanager "api/internal/monitor-manager"
)

func checkMonitorManagerMessagePayload(payload *monitormanager.MonitorManagerMessage) (bool, string) {
	if payload.Option != 0 && payload.Option != 1 {
		return false, `invalid "option" field, value must be 1 or 0`
	}
	if payload.Type != monitormanager.ICMARKETS {
		return false, `invalid "type" field`
	}
	return true, ""
}
