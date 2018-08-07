package main

import (
 "encoding/json"
  "fmt"
  "time"
  "log"
  "net/http"
  "os"

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/aws/credentials"
  "github.com/guregu/dynamo"
  "github.com/rs/xid"
  "github.com/joho/godotenv"

  "github.com/gorilla/mux"
)

type User struct {
  UserID      xid.ID    `dynamo:"user_id"`
  Name        string    `dynamo:"name"`
  CreatedTime time.Time `dynamo:"created_time"`
}

var err = godotenv.Load()
/* 関数外ではこんな記述はできないが、envをグローバルで使いたい。今のままではエラー処理ができない
 if err != nil {
     log.Fatal("Error loading .env file")
 }
*/

var cred = credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY"), os.Getenv("AWS_SECRET_KEY"), "") // 最後の引数は[セッショントークン]
var db = dynamo.New(session.New(), &aws.Config{
    Credentials: cred,
    Region: aws.String(os.Getenv("AWS_REGION")), // "ap-northeast-1"等
})
var table = db.Table("User")

func GetUsers(w http.ResponseWriter, r *http.Request) {
  var users []User
  err := table.Scan().All(&users)
  if err != nil {
    fmt.Println("err")
    panic(err.Error())
  }

  for i := range users {
    fmt.Println(users[i])
  }
  json.NewEncoder(w).Encode(users)
}

// GetUser func　id指定したユーザー情報所得
func GetUser(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    userID := params["id"]

    var user []User
    err := table.Get("user_id", userID).All(&user)
    if err != nil {
        fmt.Println("err")
        panic(err.Error())
    }
    fmt.Println(user)
    json.NewEncoder(w).Encode(user)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
  guid := xid.New()
  name := r.URL.Query().Get("name")//変更！！　作成するユーザ名をクエリから取得。
  u := User{UserID: guid, Name: name, CreatedTime: time.Now().UTC()}
  if err := table.Put(u).Run(); err != nil {
      fmt.Println("err")
      panic(err.Error())
  }
}

// UpdateUser func　id指定したユーザー情報（名前）を更新
func UpdateUser(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    q := r.URL.Query().Get("name")
    userID := params["id"]
    err := table.Update("user_id", userID).Set("name", q).Run()
    if err != nil {
        fmt.Println("err")
        panic(err.Error())
    }
}

// DeleteUser func　id指定したユーザー情報を削除
func DeleteUser(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    userID := params["id"]
    table.Delete("user_id", userID).Run()
}

func main() {
  router := mux.NewRouter()
  fmt.Println("Listening 8000 ...")
  router.HandleFunc("/users", GetUsers).Methods("GET")
  router.HandleFunc("/users/{id}", GetUser).Methods("GET")
  router.HandleFunc("/users", CreateUser).Methods("POST")
  router.HandleFunc("/users/{id}", UpdateUser).Methods("PUT")
  router.HandleFunc("/users/{id}", DeleteUser).Methods("DELETE")
  log.Fatal(http.ListenAndServe(":8000", router))
}
