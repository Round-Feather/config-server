package main

import (
	"fmt"
	"github.com/roundfeather/configuration-server/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil && err.Error() != "" {
		fmt.Println(err)
	}
}
