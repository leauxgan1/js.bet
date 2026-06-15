package eventlog


var EventLog FighterEventLog

type FighterEventLog struct {
	Log []string
}

func New() FighterEventLog {
	return FighterEventLog {
		Log: []string{},
	}
}
func NewFrom(data string) FighterEventLog {
	newLog := []string{}
	return FighterEventLog{
		Log: append(newLog, data),
	}
}

func (l *FighterEventLog) Write(data string) {
	l.Log = append(l.Log, data)
}


