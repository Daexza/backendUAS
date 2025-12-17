// package main

// import (
// 	"fmt"
// 	"golang.org/x/crypto/bcrypt"
// )

// func main() {
// 	password := "mahasiswa1" // ganti dengan password yang diinginkan
// 	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("HASH:", string(hash))
// }