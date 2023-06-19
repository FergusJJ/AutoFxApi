package ctrader

type CtraderMonitorMessage struct {
	CopyPID     string  `json:"copyPID"`
	SymbolID    int     `json:"symbolID"`
	Price       float64 `json:"price"`
	Volume      int     `json:"volume"`
	Direction   string  `json:"direction"`
	MessageType string  `json:"type"` //close or open
}
