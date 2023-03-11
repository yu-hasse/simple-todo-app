package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var Layout = "2006/01/02"

var DB *sql.DB
var todos []TodoObj = []TodoObj{}

type TodoObj struct {
	MetaData
	Todo
}

type MetaData struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
}

type Todo struct {
	Author      string    `json:"author"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"dueDate"`
}

func list(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query("SELECT * FROM todos")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	todos := make([]TodoObj, 0)
	for rows.Next() {
		var todo TodoObj
		if err := rows.Scan(&todo.ID, &todo.Author, &todo.Name, &todo.Description, &todo.DueDate, &todo.CreatedAt); err != nil {
			panic(err)
		}
		todos = append(todos, todo)
	}

	_ = json.NewEncoder(w).Encode(todos)
}

func add(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	if r.Body == http.NoBody {
		_ = encoder.Encode("request body is empty")
		return
	}

	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		_ = encoder.Encode(err.Error())
		return
	}

	obj := TodoObj{
		MetaData: MetaData{
			CreatedAt: time.Now(),
		},
		Todo: todo,
	}
	result, err := DB.Exec("INSERT INTO todos (author, name, description, due_date, created_at) VALUES (?,?,?,?,?)", obj.Author, obj.Name, obj.Description, obj.DueDate.Format("2006-01-02 15:03:04"), obj.CreatedAt.Format("2006-01-02 15:03:04"))
	if err != nil {
		panic(err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		panic(err)
	}
	var responseData TodoObj
	if err := DB.QueryRow("SELECT * FROM todos WHERE id=?", lastInsertID).Scan(&responseData.ID, &responseData.Author, &responseData.Name, &responseData.Description, &responseData.DueDate, &responseData.CreatedAt); err != nil {
		panic(err)
	}
	_ = encoder.Encode(obj)
}

func main() {
	db, err := sql.Open("mysql", "root:mysql@tcp(127.0.0.1:3306)/todo?parseTime=true")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		panic(err)
	}

	if _, err := DB.Exec("CREATE TABLE IF NOT EXISTS todos(id INTEGER AUTO_INCREMENT, author VARCHAR(32), name VARCHAR(32), description VARCHAR(64), due_date DATETIME, created_at DATETIME, PRIMARY KEY (id))"); err != nil {
		panic(err)
	}
	http.HandleFunc("/todo/list", list)
	http.HandleFunc("/todo/add", add)
	http.ListenAndServe(":8080", nil)
}
