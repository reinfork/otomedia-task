package main

import (
	"log"
	"os"

	"github.com/ucup/taskmanager/internal/model"
	"github.com/ucup/taskmanager/pkg/db"
)

// Seed data — 10 realistic tasks covering every status, several assignees,
// and titles/descriptions that exercise keyword search nicely.
var seedTasks = []model.Task{
	{Title: "Fix login page crash on empty password", Description: "App crashes when user submits login form with empty password field", Status: model.StatusTodo, Assignee: "budi"},
	{Title: "Implement email verification flow", Description: "Send verification link after register, resend token endpoint", Status: model.StatusInProgress, Assignee: "sari"},
	{Title: "Write API documentation for v1", Description: "Document all /api/tasks endpoints with request/response examples", Status: model.StatusTodo, Assignee: "dewi"},
	{Title: "Add pagination to reports module", Description: "Reports query loads 50k rows at once, needs page/limit", Status: model.StatusTodo, Assignee: "budi"},
	{Title: "Migrate session storage to Redis", Description: "Move in-memory session map to Redis with TTL for horizontal scaling", Status: model.StatusInProgress, Assignee: "rini"},
	{Title: "Fix duplicate invoice number generation", Description: "Race condition in invoice counter produces duplicates under load", Status: model.StatusDone, Assignee: "budi"},
	{Title: "Design dark mode color palette", Description: "Prepare token set for dark theme, focus on contrast ratios", Status: model.StatusTodo, Assignee: "dewi"},
	{Title: "Set up staging environment CI pipeline", Description: "GitHub Actions: lint, test, build, deploy to staging on main merge", Status: model.StatusDone, Assignee: "rini"},
	{Title: "Improve search relevance on task title", Description: "Rank exact matches above partial ILIKE matches", Status: model.StatusInProgress, Assignee: "sari"},
	{Title: "Write onboarding guide for new devs", Description: "README walkthrough: setup, migration, seed, run, test", Status: model.StatusTodo, Assignee: "dewi"},
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	_ = os.Setenv("DATABASE_URL", getEnv("DATABASE_URL", "taskuser:taskpass@tcp(localhost:3306)/taskdb?parseTime=true&charset=utf8mb4&loc=Local"))
	dsn := os.Getenv("DATABASE_URL")

	database, err := db.ConnectMySQL(dsn)
	if err != nil {
		log.Fatalf("failed to connect mysql: %v", err)
	}
	if err := db.AutoMigrate(database); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	var existing int64
	if err := database.Model(&model.Task{}).Count(&existing).Error; err != nil {
		log.Fatalf("failed to count tasks: %v", err)
	}

	created := 0
	for _, t := range seedTasks {
		// Idempotent: skip titles that already exist (including soft-deleted).
		var dup int64
		database.Unscoped().Model(&model.Task{}).Where("title = ?", t.Title).Count(&dup)
		if dup > 0 {
			continue
		}
		if err := database.Create(&t).Error; err != nil {
			log.Printf("skip %q: %v", t.Title, err)
			continue
		}
		created++
	}
	log.Printf("seed done: %d created, %d skipped (already existed), %d rows total before seed", created, len(seedTasks)-created, existing)
}
