package dao

type TodoList struct {
	Meta
	UserId      string
	Name        string
	Description string
}

type AggregateTodoList struct {
	TodoList
	Items          []TodoItem
	TotalItems     int
	CompletedItems int
}
