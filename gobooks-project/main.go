package main

import (
	"errors"
	"fmt"
	"gobooks/internal/service"
)

func soma(x, y int) (int, error) {
	result := x + y
	return result, nil
}

func div(x, y int) (int, error) {
	if x == 0 || y == 0 {
		return 0, errors.New("divisão por zero")
	}
	result := x / y
	return result, nil
}

func main() {
	fmt.Println("Hello", "World", 2030)
	x, _ := soma(10, 20)
	fmt.Println(x)
	x, _ = div(0, 0)
	fmt.Println(x)

	book := service.Book{
		ID:     1,
		Title:  "The Hobbit",
		Author: "J. R. R. Tolkien",
		Genre:  "Fantasy",
	}

	fmt.Println(book.GetFullBook())

}
