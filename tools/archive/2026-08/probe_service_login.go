package main
import (
  "context"
  "fmt"
  "github.com/jmoiron/sqlx"
  "github.com/partflow/smart-store/internal/auth"
  _ "modernc.org/sqlite"
)
func main(){
  db, err := sqlx.Connect("sqlite", `C:\Users\Administrator\Desktop\PartFlow\local-partflow.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)`)
  if err != nil { panic(err) }
  defer db.Close()
  svc, err := auth.NewService(db, "test-secret", false, "", "")
  if err != nil { panic(err) }
  res, err := svc.Login(context.Background(), &auth.LoginRequest{Email:"owner@partflow.com", Password:"Owner123456"})
  fmt.Printf("res=%#v\nerr=%v\n", res, err)
}
