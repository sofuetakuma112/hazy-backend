package main

import (
	"encoding/json"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
)

type Message struct {
	Message string
}

type User struct {
	gorm.Model
	Name  string
	Email string
}

func handleTest(w http.ResponseWriter, r *http.Request) {
	message := Message{
		Message: "this is test",
	}
	output, err := json.MarshalIndent(&message, "", "\t\t")
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-type", "application/json")
	w.Write(output)
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	user := User{}
	db.First(&user, 1)
	output, err := json.MarshalIndent(&user, "", "\t\t")
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-type", "application/json")
	w.Write(output)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	db.Create(&User{Name: "Takuma Sofue", Email: "kawahagi0620@gmail.com"})
	message := Message{
		Message: "success!",
	}
	output, err := json.MarshalIndent(&message, "", "\t\t")
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-type", "application/json")
	w.Write(output)
}

var db *gorm.DB

func init() {
	var err error
	dsn := "host=postgresql user=develop password=develop dbname=develop port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
}

func main() {
	db.AutoMigrate(&User{}) // Migrate the schema

	server := http.Server{
		Addr: ":8080",
	}
	http.HandleFunc("/test", handleTest)
	http.HandleFunc("/getUser", handleGetUser)
	http.HandleFunc("/createUser", handleCreateUser)
	server.ListenAndServe()
}
