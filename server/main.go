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
	json_bytes, err := json.MarshalIndent(&message, "", "\t\t")
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-type", "application/json")
	w.Write(json_bytes)
}

func handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 400)
		return
	}
	jwt := cookie.Value

	claims, err := verificateJWT(jwt) // JWTの署名を検証してOKならclaimsを返す
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 400)
		return
	}

	var user User
	db.First(&user, claims["user_id"])
	w.Header().Set("Content-type", "application/json")
	json_bytes, _ := json.MarshalIndent(&user, "", "\t\t")
	w.Write(json_bytes)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		len := r.ContentLength
		body := make([]byte, len)
		r.Body.Read(body)
		// TODO: JSONデータのキーが正しいかチェックする必要がある
		m := make(map[string]interface{})
		err := json.Unmarshal(body, &m)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), 400)
			return
		}
		fmt.Println(m)

		var user User
		if err := db.Where("email = ? AND password = ?", m["email"], m["password"]).First(&user).Error; err != nil {
			http.Error(w, fmt.Sprintf("%v", err), 400)
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
		json_bytes, _ := json.MarshalIndent(&user, "", "\t\t")
		w.Write(json_bytes)
	}
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
		json_bytes, err := json.MarshalIndent(&user, "", "\t\t")
		if err != nil {
			panic(err)
		}
		w.Write(json_bytes)
	} else {
		// 対応していないHttpメソッドなのでエラーを返す
	}
}

func authMiddleWare(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// JWTが有効かチェックする
		cookie, err := r.Cookie("jwt")
		if err != nil {
			// クッキーが無い時のエラー？
			http.Error(w, fmt.Sprintf("%v", err), 400)
			return
		}
		jwt := cookie.Value

		_, err = verificateJWT(jwt) // JWTの署名を検証してOKならclaimsを返す
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), 401)
			return
		}
		f(w, r)
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
	http.HandleFunc("/getCurrentUser", handleGetCurrentUser)
	http.HandleFunc("/createUser", handleCreateUser)
	http.HandleFunc("/login", handleLogin)
	server.ListenAndServe()
}
