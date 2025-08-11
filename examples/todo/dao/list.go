package dao

type TodoList struct {
	Meta
	UserId      string
	Name        string
	Description string
	Items       []TodoItem
}

type AggregateTodoList struct {
	TodoList
	TotalItems     int
	CompletedItems int
}
