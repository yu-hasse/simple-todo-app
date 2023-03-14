package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var gormDB *gorm.DB

var redisClient *redis.Client

type Todo struct {
	gorm.Model
	UserID      int       `json:"userID"`
	Author      string    `json:"author"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"dueDate"`
}

func list(c *gin.Context) {
	userID := c.Param("user_id")

	todos := make([]Todo, 0)

	// 最初にredisから取得する
	if b, err := redisClient.Get(c, userID).Bytes(); err == nil {
		if err := json.Unmarshal(b, &todos); err == nil {
			c.JSON(http.StatusOK, todos)
			return
		}
	}

	if err := gormDB.WithContext(c).Find(&todos, "user_id = ?", userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	j, err := json.Marshal(todos)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := redisClient.Set(c, userID, j, 5*time.Minute).Err(); err != nil {
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

	if err := redisClient.Del(c, strconv.Itoa(todo.UserID)).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, todo)
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/todos/:user_id", list)
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

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	log.Println(setupRouter().Run(":8080"))
}
