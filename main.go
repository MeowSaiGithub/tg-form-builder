package main

import (
	"go-tg-support-ticket/cmd"
	_ "go-tg-support-ticket/internal/database/mongo"
	_ "go-tg-support-ticket/internal/database/mysql"
	_ "go-tg-support-ticket/internal/database/postgres"
	_ "go-tg-support-ticket/internal/database/sqlite"
)

func main() {
	cmd.Execute()
}
