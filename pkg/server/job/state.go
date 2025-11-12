package job

type State int

const (
	StateUnknown   = -1
	StateSuccess   = 0
	StateRunning   = 1
	StateError     = 2
	StateWait      = 3
	StateTaskLimit = 4
)

// ResponseJob state strings (Japanese)
const (
	StateStringWait      = "開始待ち"
	StateStringRunning   = "実行中"
	StateStringSuccess   = "終了"
	StateStringError     = "エラー"
	StateStringTaskLimit = "Limit"
	StateStringUnknown   = "Unknown State"
)

func (s State) String() string {
	switch s {
	case StateSuccess:
		return "SUCCESS"
	case StateRunning:
		return "Running"
	case StateError:
		return "ERROR"
	case StateWait:
		return "WAIT"
	case StateTaskLimit:
		return "TaskLimit"
	default:
		return "Unknown State"
	}
}
