package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Message struct {
	Message string
}

type User struct {
	gorm.Model
	Name     string
	Email    string `gorm:"unique"`
	Password string
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

func handleLogin(w http.ResponseWriter, r *http.Request) {
	//
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		len := r.ContentLength
		body := make([]byte, len)
		r.Body.Read(body)
		var user User
		json.Unmarshal(body, &user)
		result := db.Create(&user) // DBに保存する
		if result.Error != nil {
			// emailの重複エラーはここで検知できる
			http.Error(w, fmt.Sprintf("%v", result.Error), 400)
			return
		}
		// user.IDを元にJWTを生成する
		jwt := generateJWT(int(user.ID))
		cookie := &http.Cookie{
			Name:  "jwt", // ここにcookieの名前を記述
			Value: jwt,   // ここにcookieの値を記述
		}
		http.SetCookie(w, cookie)
		w.Header().Set("Content-type", "application/json")
		output, err := json.MarshalIndent(&user, "", "\t\t")
		if err != nil {
			panic(err)
		}
		w.Write(output)
	} else {
		// 対応していないHttpメソッドなのでエラーを返す
	}
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
	db.Exec("DELETE FROM users")
	db.AutoMigrate(&User{}) // Migrate the schema

	server := http.Server{
		Addr: ":8080",
	}
	http.HandleFunc("/test", handleTest)
	http.HandleFunc("/getUser", handleGetUser)
	http.HandleFunc("/createUser", handleCreateUser)
	server.ListenAndServe()

	// jwt := generateJWT(1)
	// claims, err := verificateJWT(jwt)
	// if err != nil {
	// 	fmt.Println(err)
	// 	// http.Error(w, fmt.Sprintf("...: %w", err) , 400)
	// }
	// fmt.Println(claims)
}
