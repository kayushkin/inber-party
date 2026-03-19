package main

import (
	"fmt"
	"github.com/kayushkin/inber-party/internal/inber"
)

func main() {
	examples := []struct{
		input  string
		status string
	}{
		{"Spawn brigid to add a comment at the top of src/pages/Home.tsx saying // Updated by Brigid via gateway. Just add the comment, nothing else. Report when done.", "completed"},
		{"cd to /home/slava/life/repos/kayushkin.com and add a comment at the very top of src/pages/Home.tsx saying: // Updated by Brigid via gateway.", "success"},
		{"have brigid add something fun to the homepage", "completed"},
		{"Can you have brigid add one more easter egg", "completed"},
		{"Fix the broken authentication system", "error"},
		{"Build a new user dashboard with metrics", "running"},
		{"Deploy the latest version to production", "completed"},
		{"Refactor the database layer for better performance", "running"},
		{"Write documentation for the API endpoints", "completed"},
		{"Debug the memory leak in the worker process", "failed"},
	}
	
	for i, ex := range examples {
		fmt.Printf("Example %d:\n", i+1)
		fmt.Printf("Input: %s\n", ex.input)
		fmt.Printf("Status: %s\n", ex.status)
		// We can't directly call the internal function, so let's just show the input/status
		fmt.Printf("---\n")
	}
}
