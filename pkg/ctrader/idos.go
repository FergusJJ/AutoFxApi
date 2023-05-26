package ctrader

type CtraderMonitorMessage struct {
	Symbol  string `json:"symbol"`
	Message string `json:"message"`
}
