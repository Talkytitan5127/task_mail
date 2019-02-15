package room

type Room struct {
	users map[string]bool
	messages []string
}

func (r Room) is_user_in_room(name string) bool {
	return false
}

func (r Room) get_messages() []string {
	res := make([]string, 5)
	return res
}