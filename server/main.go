package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
)

type User struct {
	gorm.Model
	Name     string
	Email    string `gorm:"unique"`
	Password string
	Friends  []*User `gorm:"many2many:user_friends"` // many2manyで指定した名前の中間テーブルが作成される
}

type UserFriend struct {
	UserID   int `gorm:"primaryKey"`
	FriendID int `gorm:"primaryKey"`
	// IsFriend  bool
	// IsFiled   bool
	// IsBlocked bool
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
	if err = db.First(&user, claims["user_id"]).Error; err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 400)
		return
	}
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
		m := make(map[string]string)
		err := json.Unmarshal(body, &m)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), 400)
			return
		}

		var user User
		if err := db.Where("email = ?", m["email"]).First(&user).Error; err != nil {
			http.Error(w, fmt.Sprintf("%v", err), 400)
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(m["password"]))
		if err != nil {
			// パスワードが違う
			http.Error(w, fmt.Sprintf("%v", "Email or password is wrong"), 400)
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

func handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt")
	if err == http.ErrNoCookie {
		http.Error(w, fmt.Sprintf("%v", err), 400)
	}
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
	fmt.Fprint(w, "success!")
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		len := r.ContentLength
		body := make([]byte, len)
		r.Body.Read(body)

		// TODO: 文字列型以外のデータが送られてきたときにエラーになるかも？
		m := make(map[string]string)
		json.Unmarshal(body, &m)

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(m["password"]), 10)
		if err != nil {
			fmt.Println(err)
			return
		}

		user := User{
			Name:     m["name"],
			Email:    m["email"],
			Password: string(hashedPassword),
		}
		if err := db.Create(&user).Error; err != nil {
			// emailの重複エラーはここで検知できる
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
		json_bytes, err := json.MarshalIndent(&user, "", "\t\t")
		if err != nil {
			panic(err)
		}
		w.Write(json_bytes)
	} else {
		// 対応していないHttpメソッドなのでエラーを返す
	}
}

func handleGetFriends(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("jwt")
	jwt := cookie.Value
	claims, _ := verificateJWT(jwt)
	var user User
	if err := db.First(&user, claims["user_id"]).Error; err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 400)
		return
	}
	var friends []User
	if err := db.Model(&user).Association("Friends").Find(&friends); err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 400)
		return
	}

	// friendsのIDをUserIDとして、userのIDをFriendIDとして中間テーブルを検索する(ヒットしたfriendsのIDが相互フォロー)
	// FriendsフィールドのUserスライスからIDのみを抽出したスライスを作成する
	var userIds_friend []int
	for _, friend := range friends {
		userIds_friend = append(userIds_friend, int(friend.ID))
	}
	// FriendのUserが、友達リストを取得しようとしているUserをフォローしているレコードを中間テーブルから取得
	var userFriends []UserFriend
	if err := db.Model(&UserFriend{}).Where("user_id in (?) AND friend_id = ?", userIds_friend, user.ID).Find(&userFriends).Error; err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 400)
		return
	}

	// 友達リストを取得しようとしているUserをフォローしているFriendUserのIDをスライスに抽出する
	var friendUserIds []int // 相互フォローのFriendsフィールドのuids
	for _, friend := range userFriends {
		friendUserIds = append(friendUserIds, friend.UserID)
	}

	// friendUserIdsを元に、友達リストを取得しようとしているUserが申請中のUserと、友達のUserに振り分ける
	var friendsWithRelationStatus []interface{}
	for _, friend := range friends {
		isFriend := false
		for _, userId := range friendUserIds {
			if int(friend.ID) == userId {
				isFriend = true
			}
		}
		friendsWithRelationStatus = append(friendsWithRelationStatus, map[string]interface{}{
			"userId":         friend.ID,
			"name":           friend.Name,
			"isFriend": isFriend,
		})
	}

	json_map := map[string]interface{}{
		"friendsWithRelationStatus": friendsWithRelationStatus,
	}

	json_bytes, err := json.MarshalIndent(&json_map, "", "\t\t")
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-type", "application/json")
	w.Write(json_bytes)
}

func handleAddFriend(w http.ResponseWriter, r *http.Request) {
	// 以下のようなJSONを想定している
	// {
	// 	"userId": 2
	// }
	length := r.ContentLength
	body := make([]byte, length)
	r.Body.Read(body)

	m := make(map[string]int)
	json.Unmarshal(body, &m)

	userId := m["userId"]
	var userToAddFriend User
	if err := db.First(&userToAddFriend, userId).Error; err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 400)
		return
	}

	// authMiddleWareを通過しているのでエラーハンドリングする必要がない
	cookie, _ := r.Cookie("jwt")
	jwt := cookie.Value
	claims, _ := verificateJWT(jwt)
	var user User
	if err := db.First(&user, claims["user_id"]).Error; err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 400)
		return
	}
	// 既にuserIdとFriendIdの組み合わせが、中間テーブルに存在するかチェックする
	// TODO: 相手が自分に友達申請を出している場合に対応する
	count_myself := db.Model(&user).Where("friend_id = ?", userToAddFriend.ID).Association("Friends").Count() // 自分が対象のユーザーを友達申請していないか
	// count_opponent := db.Model(&userToAddFriend).Where("friend_id = ?", user.ID).Association("Friends").Count() // 対象のユーザーが自分を友達申請していないか
	if count_myself != 0 {
		// 既にフォローしているので何もしない
		return
	}
	// 中間テーブル(user_friends)にはuser_idが&user.ID, friend_idが&userToAddFriend.IDで登録される
	if err := db.Model(&user).Association("Friends").Append(&userToAddFriend); err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 400)
		return
	}
	fmt.Fprint(w, "success!")
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
	db.Exec("DROP TABLE IF EXISTS user_friends")
	db.Exec("DROP TABLE IF EXISTS users")
	db.AutoMigrate(&UserFriend{})
	db.AutoMigrate(&User{}) // Migrate the schema
	for _, value := range []int{1, 2, 3, 4, 5} {
		name := fmt.Sprintf("sofue%d", value)
		email := fmt.Sprintf("kawahagi0620+%d@gmail.com", value)
		password_plain := fmt.Sprintf("test%d", value)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password_plain), 10)
		if err != nil {
			panic(err)
		}

		var user = User{Name: name, Email: email, Password: string(hashedPassword)}
		db.Create(&user)
	}
	var userFriend = UserFriend{UserID: 1, FriendID: 2}
	db.Create(&userFriend)
	userFriend = UserFriend{UserID: 1, FriendID: 4}
	db.Create(&userFriend)
	userFriend = UserFriend{UserID: 2, FriendID: 3}
	db.Create(&userFriend)
	userFriend = UserFriend{UserID: 2, FriendID: 1}
	db.Create(&userFriend)
	userFriend = UserFriend{UserID: 3, FriendID: 5}
	db.Create(&userFriend)

	// UserモデルのFriendフィールドの結合テーブルをUserFriendに変更する
	// UserFriendには必要な外部キーが全て定義されていなければならず、定義されていない場合はエラーとなる
	if err := db.SetupJoinTable(&User{}, "Friends", &UserFriend{}); err != nil {
		panic(err)
	}

	server := http.Server{
		Addr: ":8080",
	}
	http.HandleFunc("/getCurrentUser", handleGetCurrentUser)
	http.HandleFunc("/createUser", handleCreateUser)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/logout", handleLogout)

	http.HandleFunc("/addFriend", handleAddFriend)
	http.HandleFunc("/getFriends", handleGetFriends)
	server.ListenAndServe()
}
