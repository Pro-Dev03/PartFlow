package main
import (
  "fmt"
  "github.com/jmoiron/sqlx"
  _ "modernc.org/sqlite"
)
func main() {
  db, err := sqlx.Connect("sqlite", `C:\Users\Administrator\Desktop\PartFlow\local-partflow.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)`)
  if err != nil { panic(err) }
  defer db.Close()
  type User struct { Email string `db:"email"`; IsActive int `db:"is_active"`; SubscriptionStatus string `db:"subscription_status"` }
  var user User
  err = db.Get(&user, `SELECT email, is_active, subscription_status FROM users WHERE email = $1 AND is_active = TRUE`, "owner@partflow.com")
  fmt.Printf("err=%v user=%+v\n", err, user)
  err = db.Get(&user, `SELECT email, is_active, subscription_status FROM users WHERE email = ? AND is_active = 1`, "owner@partflow.com")
  fmt.Printf("q2 err=%v user=%+v\n", err, user)
}
