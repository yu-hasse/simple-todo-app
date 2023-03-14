package main

import (
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

	if err := mock.ExpectationsWereMet(); err != nil{
		t.Fatal(err.Error())
	}

	response := recoder.Result()
	defer response.Body.Close()

	var responseTodos []Todo
	if err := json.NewDecoder(response.Body).Decode(&responseTodos); err != nil{
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

}
