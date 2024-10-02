package main

import (
	"context"
	"fmt"

	"github.com/jantytgat/go-jobs/pkg/library"
)

func main() {
	t := library.PrintTask{Message: "Hello!"}
	status, err := t.DefaultHandler().Execute(context.Background(), t, nil)
	fmt.Printf("Status: %s, Error: %v\r\n", status, err)
}
