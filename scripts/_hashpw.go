package main
import (
  "fmt"
  "github.com/Jiang-Xia/blog-server-go/pkg/crypto"
)
func main() {
  h, err := crypto.Hash("123456")
  if err != nil { panic(err) }
  fmt.Print(h)
}
