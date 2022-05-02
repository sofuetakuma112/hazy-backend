package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

func generateJWT(userId int) string {
	// 構造体Claimの作成
	claims := jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	// ヘッダーとペイロードを作成する
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // claimsを元にHeaderとPayloadを作成
	// fmt.Printf("Header:%#v\n", token.Header)                   // Header:map[string]interface {}{"alg":"HS256", "typ":"JWT"}
	// fmt.Printf("Claims:%#v\n", token.Claims)                   // Claims:jwt.MapClaims{"exp":1651564567, "user_id":12345678}

	tokenString, _ := token.SignedString([]byte("SECRET_KEY")) // 署名付きトークンを作成
	// fmt.Println("tokenString:", tokenString)
	return tokenString
}

func verificateJWT(tokenString string) (jwt.MapClaims, error) {
	// tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NTE1NjQ1NjcsInVzZXJfaWQiOjEyMzQ1Njc4fQ.bDrEs1r_iXRnPjTLsyxAkgddoVCdVMiFq023ohsmS5Q"

	// JWTの改ざん検知の検証は以下の流れで行う(以下の手続きをjwt.Parseが行う)
	// 1. 受け取ったトークンからヘッダーとペイロードを取り出す
	// 2. 取り出したヘッダーとペイロードから改めて署名を作成
	// 3. 作成した署名と、トークンに含まれる署名を比較
	// 4. 一致していれば正しいトークン、不一致であれば正しくないトークン

	// jwt.Parseは第一引数に検証するトークン、第二引数に暗号化に用いたキーを検索するための関数を渡す
	// 実行することでトークン文字列からTokenオブジェクトへとパースするとともに、内部的に検証を行い、Token.Validに検証結果を格納する
	// jwt.Parseの返り値に検証結果をValidに格納したTokenを返す
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("SECRET_KEY"), nil
	})
	if err != nil {
		return nil, err
	}

	// ペイロードが正しく取り出せて、かつトークンの検証が成功した場合はペイロードの中身を表示する
	// token.Claims.(jwt.MapClaims)はtoken.Claimsの値をjwt.MapClaimsにタイプコンバージョンしてる？？
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("Error: %s", "no claims")
	}
	if !token.Valid {
		// Tokenの検証に失敗したことをエラーで返す
		return nil, fmt.Errorf("Error: %s", "Invalid token")
	}
	return claims, nil
}
