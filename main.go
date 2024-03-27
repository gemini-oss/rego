/*
# Main

:Copyright: (c) 2024 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// cmd/main.go
package main

import (
	"github.com/gemini-oss/rego/pkg/common/log"
)

func main() {
	l := log.NewLogger("{main}", log.DEBUG)

	l.Println("Starting application...")

	// Initialize clients here

	l.Println("Application started.")
	l.Println("---------------------")

	// Build custom logic here
}
