package exceptions

import (
	"fmt"
	"log"
)

func CheckError(err error, errMsg string) {
	if err != nil {
		log.Printf(" %v", errMsg)
		err := fmt.Errorf("error : %w", err)
		_ = fmt.Sprintf("Error initializing Langchain client: %v", err)
		panic(err)
	}
}

func RecoverFromError() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error encountered:", r)
			fmt.Println("The program will exit now.")
		}
	}()
}