package main
import (
  "database/sql"
  "fmt"
  _ "github.com/mattn/go-sqlite3"
)
func main(){
  path := `C:\Users\Administrator\AppData\Roaming\PartFlow\data\partflow.db`
  db, err := sql.Open("sqlite3", path)
  if err != nil { panic(err) }
  defer db.Close()
  queries := []string{
    "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;",
    "SELECT id, product_id, quantity, reserved_quantity, created_at, updated_at FROM inventory ORDER BY created_at LIMIT 20;",
    "SELECT id, product_id, item_code, status, supplier_id, purchase_date, created_at FROM inventory_items ORDER BY created_at DESC LIMIT 20;",
    "SELECT id, purchase_id, supplier_id, reason, status, refund_amount, created_at, updated_at FROM supplier_returns ORDER BY created_at DESC LIMIT 20;",
    "SELECT id, supplier_return_id, product_id, quantity, unit_cost, inventory_item_id FROM supplier_return_items ORDER BY created_at DESC LIMIT 20;",
  }
  for _, q := range queries {
    fmt.Println("\n-- QUERY --")
    fmt.Println(q)
    rows, err := db.Query(q)
    if err != nil { panic(err) }
    cols, _ := rows.Columns();
    for _, c := range cols { fmt.Printf("%s\t", c) }
    fmt.Println()
    for rows.Next(){
      vals := make([]interface{}, len(cols))
      scanArgs := make([]interface{}, len(cols))
      for i := range vals { scanArgs[i] = &vals[i] }
      if err := rows.Scan(scanArgs...); err != nil { panic(err) }
      for i:= range vals { fmt.Printf("%v\t", vals[i]) }
      fmt.Println()
    }
    rows.Close()
  }
}
