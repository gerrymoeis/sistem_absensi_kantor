package main

import (
	"fmt"
	"log"

	"absensi-app/internal/database"
	"absensi-app/internal/repository"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize database
	db, err := database.InitDB("./data/absensi.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create repository
	logRepo := repository.NewActivityLogRepository(db)

	// Get all logs
	logs, err := logRepo.FindAll(10, 0)
	if err != nil {
		log.Fatalf("Failed to get logs: %v", err)
	}

	fmt.Println("\n=== Activity Logs (Latest 10) ===")
	fmt.Println("ID | UserID | Action | Description | IP | Status | Created At")
	fmt.Println("-------------------------------------------------------------------------")

	if len(logs) == 0 {
		fmt.Println("No logs found")
	} else {
		for _, log := range logs {
			userIDStr := "NULL"
			if log.UserID != nil {
				userIDStr = fmt.Sprintf("%d", *log.UserID)
			}

			fmt.Printf("%d | %s | %s | %s | %s | %s | %s\n",
				log.ID, userIDStr, log.ActionType, log.Description,
				log.IPAddress, log.Status, log.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Printf("\nTotal logs: %d\n", len(logs))
}
