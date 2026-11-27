// Command knowledge-resync enqueues every live AI knowledge source (active products, active scripts, knowledge
// documents) of one tenant or of all tenants for sync, through trusted database access. Run it after switching the
// embedding model (AI_EMBEDDING_PROVIDER / AI_EMBEDDING_MODEL): the messaging-command sync job then re-sends each
// source and the AI service re-embeds chunks stored under another model. -status only prints job counts.
//
// The database URL comes from API_DATABASE_URL (set in the production ingress container); nothing is printed
// except counts.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var tenantID = regexp.MustCompile(`^[A-Za-z0-9-]{1,36}$`)

func main() {
	tenant := flag.String("tenant", "", "re-sync one tenant (workspace ID)")
	all := flag.Bool("all-tenants", false, "re-sync every tenant")
	status := flag.Bool("status", false, "only print knowledge sync job counts")
	flag.Parse()
	dsn := os.Getenv("API_DATABASE_URL")
	if dsn == "" {
		exit("provide API_DATABASE_URL through the trusted operator environment")
	}
	if !*status && (*all == (*tenant != "")) {
		exit("choose exactly one of -tenant <id> or -all-tenants (or -status)")
	}
	if *tenant != "" && !tenantID.MatchString(*tenant) {
		exit("invalid -tenant")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		exit("cannot connect to the API database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	repo := repository.NewKnowledgeRepository(db)
	if !*status {
		count, failure := repo.EnqueueKnowledgeResync(ctx, *tenant, time.Now().UTC())
		if failure != nil {
			exit("re-sync failed; verify the database and applied migrations")
		}
		fmt.Printf("Enqueued %d knowledge sources for sync.\n", count)
	}
	counts, failure := repo.KnowledgeSyncStatus(ctx)
	if failure != nil {
		exit("cannot read knowledge sync status")
	}
	fmt.Printf("Knowledge sync jobs: pending=%d stuck=%d failed=%d done=%d\n", counts.Pending, counts.Stuck, counts.Failed, counts.Done)
}

func exit(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
