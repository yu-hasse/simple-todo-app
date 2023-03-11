package main

import (
	"encoding/json"
	"net/http"
	"time"
)

var Layout = "2006/01/02"

var todos []TodoObj = []TodoObj{}

type TodoObj struct {
	MetaData
	Todo
}

type MetaData struct {
	ID int `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
}

type Todo struct {
	Author      string    `json:"author"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"dueDate"`
}

func list(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)

	now := time.Now().Format(Layout)
	due, err := time.Parse("2006/01/02", now)
	if err != nil {
		_ = encoder.Encode(err.Error())
		return
	}
	todo := Todo{
		Author:      "Bob",
		Name:        "Develop Generics",
		Description: "Add feature of generic type system",
		DueDate:     due,
	}
	_ = encoder.Encode(todo)
}

func add(w http.ResponseWriter, r *http.Request){
	encoder := json.NewEncoder(w)
	if r.Body == http.NoBody{
		_ = encoder.Encode("request body is empty")
		return
	}

	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo);err != nil{
		_ = encoder.Encode(err.Error())
		return
	}

	meta := MetaData{
		ID: len(todos) + 1,
		CreatedAt: time.Now(),
	}

	obj := TodoObj{
		MetaData: meta,
		Todo: todo,
	}

	todos = append(todos, obj)
	_ = encoder.Encode(obj)
}

func main() {
	http.HandleFunc("/todo/list", list)
	http.HandleFunc("/todo/add", add)
	http.ListenAndServe(":8080", nil)
}
