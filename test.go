package main

import "fmt"

type User struct {
	Name string
	Age  int
}

func get(str1 *string) {

	fmt.Println("get str=", &str1)

	*str1 = "alice"

	fmt.Println("get str2=", &str1)

}

func main() {
	str1 := "damo"
	get(&str1)

	fmt.Println("----------------")
	fmt.Println("main str=", str1)
}
