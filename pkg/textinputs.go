package pkg

import tea "github.com/charmbracelet/bubbletea"

func IsMsgNumeric(msg tea.KeyMsg) bool {
	if len(msg.Runes) != 1 {
		return false
	}
	rn := msg.Runes[0]
	return rn >= '0' && rn <= '9'
}
