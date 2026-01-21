package preflight

import (
	"claraverse/internal/database"
	"os"
	"testing"
)

func setupPreflightTest(t *testing.T) (*database.DB, func()) {
	t.Skip("SQLite tests are deprecated - preflight tests require MySQL DSN via DATABASE_URL")
	tmpDB := "test_preflight.db"

	db, err := database.New(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	if err := db.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(tmpDB)
	}

	return db, cleanup
}

func TestNewChecker(t *testing.T) {
	db, cleanup := setupPreflightTest(t)
	defer cleanup()

	checker := NewChecker(db)
	if checker == nil {
		t.Fatal("Expected non-nil checker")
	}

	if checker.db != db {
		t.Error("Checker database not set correctly")
	}
}

func TestCheckDatabaseConnection_Success(t *testing.T) {
	db, cleanup := setupPreflightTest(t)
	defer cleanup()

	checker := NewChecker(db)
	result := checker.checkDatabaseConnection()

	if result.Status != "pass" {
		t.Errorf("Expected status 'pass', got '%s'", result.Status)
	}

	if result.Name != "Database Connection" {
		t.Errorf("Expected name 'Database Connection', got '%s'", result.Name)
	}
}

func TestCheckDatabaseConnection_Failure(t *testing.T) {
	db, cleanup := setupPreflightTest(t)
	cleanup() // Close database immediately to simulate failure

	checker := NewChecker(db)
	result := checker.checkDatabaseConnection()

	if result.Status != "fail" {
		t.Errorf("Expected status 'fail', got '%s'", result.Status)
	}

	if result.Error == nil {
		t.Error("Expected error to be set")
	}
}

func TestCheckDatabaseSchema_Success(t *testing.T) {
	db, cleanup := setupPreflightTest(t)
	defer cleanup()

	checker := NewChecker(db)
	result := checker.checkDatabaseSchema()

	if result.Status != "pass" {
		t.Errorf("Expected status 'pass', got '%s': %s", result.Status, result.Message)
	}
}

func TestCheckDatabaseSchema_MissingTable(t *testing.T) {
	t.Skip("SQLite tests are deprecated - preflight tests require MySQL DSN via DATABASE_URL")
}

func TestCheckProvidersFile_Exists(t *testing.T) {
	t.Skip("Provider file checks have been removed from preflight - providers are now managed via database")
}

func TestCheckProvidersFile_Missing(t *testing.T) {
	t.Skip("Provider file checks have been removed from preflight - providers are now managed via database")
}

func TestCheckProvidersJSON_Valid(t *testing.T) {
	t.Skip("Provider file checks have been removed from preflight - providers are now managed via database")
}

func TestCheckProvidersJSON_InvalidJSON(t *testing.T) {
	t.Skip("Provider file checks have been removed from preflight - providers are now managed via database")
}

func TestCheckProvidersJSON_EmptyProviders(t *testing.T) {
	t.Skip("Provider file checks have been removed from preflight - providers are now managed via database")
}

func TestCheckProvidersJSON_MissingName(t *testing.T) {
	t.Skip("Provider file checks have been removed from preflight - providers are now managed via database")
}

func TestCheckProvidersJSON_MissingBaseURL(t *testing.T) {
	t.Skip("Provider file checks have been removed from preflight - providers are now managed via database")
}

func TestCheckProvidersJSON_MissingAPIKey(t *testing.T) {
	t.Skip("Provider file checks have been removed from preflight - providers are now managed via database")
}

func TestCheckEnvironmentVariables(t *testing.T) {
	db, cleanup := setupPreflightTest(t)
	defer cleanup()

	checker := NewChecker(db)
	result := checker.checkEnvironmentVariables()

	// Should pass or warn, but not fail
	if result.Status == "fail" {
		t.Errorf("Expected status 'pass' or 'warning', got 'fail': %s", result.Message)
	}
}

func TestRunAll(t *testing.T) {
	db, cleanup := setupPreflightTest(t)
	defer cleanup()

	checker := NewChecker(db)
	results := checker.RunAll()

	if len(results) == 0 {
		t.Error("Expected results, got empty slice")
	}

	// Verify all expected checks ran
	expectedChecks := map[string]bool{
		"Database Connection":   false,
		"Database Schema":       false,
		"Environment Variables": false,
	}

	for _, result := range results {
		if _, exists := expectedChecks[result.Name]; exists {
			expectedChecks[result.Name] = true
		}
	}

	for checkName, ran := range expectedChecks {
		if !ran {
			t.Errorf("Expected check '%s' to run", checkName)
		}
	}
}

func TestHasFailures(t *testing.T) {
	// Test with no failures
	results := []CheckResult{
		{Status: "pass"},
		{Status: "pass"},
		{Status: "warning"},
	}

	if HasFailures(results) {
		t.Error("Expected no failures")
	}

	// Test with failures
	results = append(results, CheckResult{Status: "fail"})

	if !HasFailures(results) {
		t.Error("Expected failures to be detected")
	}
}

func TestQuickCheck(t *testing.T) {
	db, cleanup := setupPreflightTest(t)
	defer cleanup()

	checker := NewChecker(db)
	results := checker.QuickCheck()

	if len(results) == 0 {
		t.Error("Expected results from quick check")
	}

	// Quick check should run fewer checks than full check
	fullResults := checker.RunAll()
	if len(results) >= len(fullResults) {
		t.Error("Expected quick check to run fewer checks than full check")
	}
}
