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
  var user auth.User
  err = db.GetContext(context.Background(), &user, `SELECT id, email, password_hash, first_name, last_name, phone, is_active, last_login_at, created_at, updated_at, subscription_status, subscription_expires_at FROM users WHERE email = ? AND is_active = TRUE`, "owner@partflow.com")
  fmt.Printf("err=%v\nuser=%+v\n", err, user)
}
