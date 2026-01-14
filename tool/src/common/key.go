package common

const (
	KEY_enter     = "enter"
	KEY_space     = " "
	KEY_backspace = "backspace"
	KEY_interrupt = "ctrl+c"
	KEY_quit      = "q"
	KEY_up        = "up"
	KEY_down      = "down"
)

var (
	modelControlKeys = []string{KEY_enter, KEY_space}
)

func ModelControlKey() []string {
	return modelControlKeys
}
