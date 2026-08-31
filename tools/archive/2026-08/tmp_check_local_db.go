package main
import (
  "database/sql"
  "fmt"
  _ "modernc.org/sqlite"
)
func main(){
  db, err := sql.Open("sqlite", `C:\Users\Administrator\AppData\Roaming\PartFlow\data\partflow.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)`)
  if err != nil { panic(err) }
  defer db.Close()
  for _, t := range []string{"users","categories","products","customers","suppliers"} {
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM " + t).Scan(&count)
    fmt.Printf("%s=%d err=%v\n", t, count, err)
  }
}
