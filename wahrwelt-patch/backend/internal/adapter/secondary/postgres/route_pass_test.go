package postgres

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"uuid"
)

func TestRoutePassPersistsOneIssue(t *testing.T) {
	dsn := os.Getenv("WAHRWELT_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("WAHRWELT_TEST_POSTGRES_DSN is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer admin.Close()
	schema := "route_pass_test_" + strings.ReplaceAll(uuid.New().String(), "-", "")
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	defer func() {
		_, _ = admin.Exec(context.Background(), "DROP SCHEMA "+quoted+" CASCADE")
	}()
	for _, statement := range []string{
		"CREATE TABLE " + quoted + ".users (id uuid PRIMARY KEY, username text NOT NULL)",
		"CREATE TABLE " + quoted + ".cases (id uuid PRIMARY KEY, owner_id uuid NOT NULL, kind text NOT NULL, origin text NOT NULL, destination text NOT NULL)",
		"CREATE TABLE " + quoted + ".route_slots (id uuid PRIMARY KEY, case_id uuid NOT NULL, owner_id uuid NOT NULL, carrier_id uuid NOT NULL, inspector_id uuid NOT NULL, slot_code text NOT NULL)",
		"CREATE TABLE " + quoted + ".route_passes (route_slot_id uuid PRIMARY KEY REFERENCES " + quoted + ".route_slots(id), issued_at timestamptz NOT NULL DEFAULT now())",
	} {
		if _, err := admin.Exec(ctx, statement); err != nil {
			t.Fatalf("set up test tables: %v", err)
		}
	}

	ownerID, carrierID, inspectorID, caseID, slotID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, user := range []struct {
		id   uuid.UUID
		name string
	}{{ownerID, "owner"}, {carrierID, "carrier"}, {inspectorID, "inspector"}} {
		if _, err := admin.Exec(ctx, "INSERT INTO "+quoted+".users (id, username) VALUES ($1, $2)", user.id, user.name); err != nil {
			t.Fatalf("insert participant: %v", err)
		}
	}
	if _, err := admin.Exec(ctx, "INSERT INTO "+quoted+".cases (id, owner_id, kind, origin, destination) VALUES ($1, $2, 'import', 'Орел', 'Тула')", caseID, ownerID); err != nil {
		t.Fatalf("insert case: %v", err)
	}
	if _, err := admin.Exec(ctx, "INSERT INTO "+quoted+".route_slots (id, case_id, owner_id, carrier_id, inspector_id, slot_code) VALUES ($1, $2, $3, $4, $5, 'private-code')", slotID, caseID, ownerID, carrierID, inspectorID); err != nil {
		t.Fatalf("insert route slot: %v", err)
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse test database: %v", err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("open isolated schema: %v", err)
	}
	defer pool.Close()
	store := &Store{pool: pool, queries: sqlc.New(pool)}

	first, err := store.CreateRoutePass(ctx, slotID)
	if err != nil {
		t.Fatalf("first issue: %v", err)
	}
	second, err := store.CreateRoutePass(ctx, slotID)
	if err != nil {
		t.Fatalf("repeat issue: %v", err)
	}
	type issueResult struct {
		issuedAt time.Time
		err      error
	}
	results := make(chan issueResult, 8)
	for i := 0; i < cap(results); i++ {
		go func() {
			issued, issueErr := store.CreateRoutePass(ctx, slotID)
			results <- issueResult{issuedAt: issued.IssuedAt, err: issueErr}
		}()
	}
	for i := 0; i < cap(results); i++ {
		result := <-results
		if result.err != nil || !result.issuedAt.Equal(first.IssuedAt) {
			t.Fatalf("concurrent issue changed the stored pass: time=%v error=%v", result.issuedAt, result.err)
		}
	}
	loaded, err := store.GetRoutePass(ctx, slotID)
	if err != nil {
		t.Fatalf("read issued pass: %v", err)
	}
	if first.SlotID != slotID || first.CaseID != caseID || first.Kind != "import" || first.Origin != "Орел" || first.Destination != "Тула" || first.CarrierUsername != "carrier" || first.InspectorUsername != "inspector" {
		t.Fatalf("issued pass has incorrect route data: %+v", first)
	}
	if first.IssuedAt.IsZero() || !first.IssuedAt.Equal(second.IssuedAt) || !first.IssuedAt.Equal(loaded.IssuedAt) {
		t.Fatalf("repeat issue changed timestamp: first=%v second=%v loaded=%v", first.IssuedAt, second.IssuedAt, loaded.IssuedAt)
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM route_passes WHERE route_slot_id=$1", slotID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("expected one stored pass, count=%d error=%v", count, err)
	}
}
