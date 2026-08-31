package main
import (
  "database/sql"
  "fmt"
  _ "modernc.org/sqlite"
)
func main() {
  db, err := sql.Open("sqlite", `C:\Users\Administrator\Desktop\PartFlow\local-partflow.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)`)
  if err != nil { panic(err) }
  defer db.Close()
  rows, err := db.Query("SELECT id,email,password_hash,first_name,last_name,is_active,subscription_status FROM users WHERE email = ?", "owner@partflow.com")
  if err != nil { panic(err) }
  defer rows.Close()
  for rows.Next() { var id,email,hash,fn,ln string; var active int; var sub string; if err:=rows.Scan(&id,&email,&hash,&fn,&ln,&active,&sub); err!=nil { panic(err)}; fmt.Printf("ROW: id=%s email=%s first=%s last=%s active=%d sub=%s hash=%s\n",id,email,fn,ln,active,sub,hash) }
  var count int; if err:=db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err!=nil { panic(err)}; fmt.Printf("COUNT=%d\n", count)
}
