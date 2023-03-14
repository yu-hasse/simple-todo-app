package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		return nil, nil, err
	}

	gormdb, err := gorm.Open(mysql.Dialector{
		Config: &mysql.Config{
			DriverName:                "mysql",
			Conn:                      sqlDB,
			SkipInitializeWithVersion: true,
		},
	}, &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	return gormdb, mock, nil
}

func TestList(t *testing.T) {
	_gormDB, mock, err := setupDB()
	if err != nil {
		t.Fatal(err.Error())
	}

	gormDB = _gormDB

	expect := Todo{
		Author:      "",
		Name:        "",
		Description: "",
		DueDate:     time.Now(),
		Model:       gorm.Model{ID: 1},
	}
	mock.ExpectQuery("SELECT * FROM `todos` WHERE `todos`.`deleted_at` IS NULL").WillReturnRows(sqlmock.NewRows([]string{"id", "author", "name", "description", "due_date"}).AddRow(expect.ID, expect.Author, expect.Name, expect.Description, expect.DueDate))

	request := httptest.NewRequest(http.MethodGet, "http://localhost:8080", nil)
	recoder := httptest.NewRecorder()
	setupRouter().ServeHTTP(recoder, request)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err.Error())
	}

	response := recoder.Result()
	defer response.Body.Close()

	var responseTodos []Todo
	if err := json.NewDecoder(response.Body).Decode(&responseTodos); err != nil {
		t.Fatal(err.Error())
	}

	if len(responseTodos) != 1 {
		t.Errorf("got response count = %d. want = %d", len(responseTodos), 1)
	}

	if responseTodos[0].ID != expect.ID {
		t.Errorf("got ID = %d. want = %d", responseTodos[0].ID, expect.ID)
	}
	if responseTodos[0].Author != expect.Author {
		t.Errorf("got Author = %s. want = %s", responseTodos[0].Author, expect.Author)
	}
	if responseTodos[0].Name != expect.Name {
		t.Errorf("got Name = %s. want = %s", responseTodos[0].Name, expect.Name)
	}
	if responseTodos[0].Description != expect.Description {
		t.Errorf("got Description = %s. want = %s", responseTodos[0].Description, expect.Description)
	}
	if responseTodos[0].DueDate.Unix() != expect.DueDate.Unix() {
		t.Errorf("got DueDate = %d. want = %d", responseTodos[0].DueDate.Unix(), expect.DueDate.Unix())
	}
}

func TestAdd(t *testing.T) {
	_gormDB, mock, err := setupDB()
	if err != nil {
		t.Fatal(err.Error())
	}
	gormDB = _gormDB

	expect := Todo{
		Author:      "author",
		Name:        "name",
		Description: "description",
		DueDate:     time.Now(),
		Model:       gorm.Model{ID: 1},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `todos` (`created_at`,`updated_at`,`deleted_at`,`author`,`name`,`description`,`due_date`,`id`) VALUES (?,?,?,?,?,?,?,?)").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), expect.Author, expect.Name, expect.Description, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	b, err := json.Marshal(expect)
	if err != nil {
		t.Fatal(err.Error())
	}

	request := httptest.NewRequest(http.MethodPost, "http://localhhost:8080/todo", bytes.NewBuffer(b))
	recoder := httptest.NewRecorder()
	setupRouter().ServeHTTP(recoder, request)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err.Error())
	}

	response := recoder.Result()
	defer response.Body.Close()

	var responseTodo Todo
	if err := json.NewDecoder(response.Body).Decode(&responseTodo); err != nil {
		t.Fatal(err.Error())
	}

	if responseTodo.ID != expect.ID {
		t.Errorf("got ID = %d. want = %d", responseTodo.ID, expect.ID)
	}
	if responseTodo.Author != expect.Author {
		t.Errorf("got Author = %s. want = %s", responseTodo.Author, expect.Author)
	}
	if responseTodo.Name != expect.Name {
		t.Errorf("got Name = %s. want = %s", responseTodo.Name, expect.Name)
	}
	if responseTodo.Description != expect.Description {
		t.Errorf("got Description = %s. want = %s", responseTodo.Description, expect.Description)
	}
	if responseTodo.DueDate.Unix() != expect.DueDate.Unix() {
		t.Errorf("got Description = %d. want = %d", responseTodo.DueDate.Unix(), expect.DueDate.Unix())
	}

}
