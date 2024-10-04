package main

import (
	"context"
	"fmt"

	"github.com/jantytgat/go-jobs/pkg/taskLibrary"
)

func main() {
	t := taskLibrary.PrintTask{Message: "Hello!"}
	status, err := t.DefaultHandler().Execute(context.Background(), t, nil)
	fmt.Printf("Status: %s, Error: %v\r\n", status, err)
}
