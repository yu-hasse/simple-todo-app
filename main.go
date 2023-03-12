package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var gormDB *gorm.DB

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
	gorm.Model
	Author      string    `json:"author"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"dueDate"`
}

func list(c *gin.Context) {
	todos := make([]Todo, 0)

	if err := gormDB.WithContext(c).Find(&todos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, todos)
}

func add(c *gin.Context) {
	if c.Request.Body == http.NoBody {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "body is required"})
		return
	}

	var todo Todo
	if err := json.NewDecoder(c.Request.Body).Decode(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body is invalid"})
		return
	}

	if err := gormDB.WithContext(c).Create(&todo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, todo)
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/todos", list)
	router.POST("/todo", add)
	return router
}

func main() {
	_gormDB, err := gorm.Open(mysql.Open("root:mysql@tcp(127.0.0.1:3306)/todo?parseTime=true"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db, err := _gormDB.DB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if !_gormDB.Migrator().HasTable(&Todo{}) {
		if err := _gormDB.Migrator().CreateTable(&Todo{}); err != nil {
			panic(err)
		}
	}

	gormDB = _gormDB

	log.Println(setupRouter().Run(":8080"))
}
