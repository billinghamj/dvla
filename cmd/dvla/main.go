package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/YodaTheCoder/dvla"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "dvla: provide a vrm")
		return
	}

	details, err := dvla.Check(os.Args[1])
	if err != nil {
		panic(err)
	}

	data, err := json.Marshal(details)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
}
