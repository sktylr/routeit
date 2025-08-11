package dao

type TodoList struct {
	Meta
	UserId      string
	Name        string
	Description string
	Items       []TodoItem
}
