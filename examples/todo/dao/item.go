package dao

type TodoItem struct {
	Meta
	UserId     string
	TodoListId string
	Name       string
	Status     string
}
