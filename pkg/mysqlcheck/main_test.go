package mysqlcheck

import (
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/tapclap/db-connect-checker/pkg/types"
)

func TestGetSQLTables(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(sqlmock.Sqlmock)
		expectedCount int
		expectedTable []string
		wantErr       bool
		errContains   string
	}{
		{
			name: "successfully retrieves tables",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"Tables_in_db"}).
					AddRow("users").
					AddRow("products").
					AddRow("orders")
				mock.ExpectQuery("SHOW TABLES").WillReturnRows(rows)
			},
			expectedCount: 3,
			expectedTable: []string{"users", "products", "orders"},
			wantErr:       false,
		},
		{
			name: "successfully retrieves single table",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"Tables_in_db"}).
					AddRow("users")
				mock.ExpectQuery("SHOW TABLES").WillReturnRows(rows)
			},
			expectedCount: 1,
			expectedTable: []string{"users"},
			wantErr:       false,
		},
		{
			name: "successfully retrieves empty table list",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"Tables_in_db"})
				mock.ExpectQuery("SHOW TABLES").WillReturnRows(rows)
			},
			expectedCount: 0,
			expectedTable: []string{},
			wantErr:       false,
		},
		{
			name: "returns error when query fails",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SHOW TABLES").
					WillReturnError(errors.New("connection lost"))
			},
			wantErr:     true,
			errContains: "connection lost",
		},
		{
			name: "returns error when database is locked",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SHOW TABLES").
					WillReturnError(errors.New("database is locked"))
			},
			wantErr:     true,
			errContains: "database is locked",
		},
		{
			name: "handles many tables",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"Tables_in_db"})
				expectedTables := make([]string, 100)
				for i := 0; i < 100; i++ {
					tableName := fmt.Sprintf("table_%d", i)
					rows.AddRow(tableName)
					expectedTables[i] = tableName
				}
				mock.ExpectQuery("SHOW TABLES").WillReturnRows(rows)
			},
			expectedCount: 100,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DB
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create sqlmock: %v", err)
			}
			defer db.Close()

			// Setup mock expectations
			tt.mockSetup(mock)

			// Execute function
			tables, err := getSQLTables(db)

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Error("getSQLTables() expected error but got none")
				}
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("getSQLTables() error = %v, want error containing %v", err, tt.errContains)
					}
				}
			} else {
				if err != nil {
					t.Errorf("getSQLTables() unexpected error: %v", err)
				}
				if len(tables) != tt.expectedCount {
					t.Errorf("getSQLTables() returned %d tables, want %d", len(tables), tt.expectedCount)
				}
				// Check specific table names if provided
				if tt.expectedTable != nil && len(tt.expectedTable) > 0 {
					for i, expectedName := range tt.expectedTable {
						if i >= len(tables) {
							t.Errorf("getSQLTables() missing expected table at index %d: %s", i, expectedName)
							break
						}
						if tables[i] != expectedName {
							t.Errorf("getSQLTables() table[%d] = %s, want %s", i, tables[i], expectedName)
						}
					}
				}
			}

			// Verify all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestCheckConnection(t *testing.T) {
	tests := []struct {
		name      string
		config    types.MysqlConfig
		mockSetup func(sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "successful connection without TLS",
			config: types.MysqlConfig{
				Name: "testdb",
				User: "testuser",
				Pass: "testpass",
				Host: "localhost",
				Port: "3306",
				TLS:  false,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"Tables_in_db"}).
					AddRow("users")
				mock.ExpectQuery("SHOW TABLES").WillReturnRows(rows)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Testing checkMysqlConnection with real DB connection is complex
			// because sql.Open creates a real connection pool.
			// For unit tests, we would need to refactor the code to accept *sql.DB
			// as a parameter or use an interface.

			// This test is more of an integration test and would require
			// either a real MySQL instance or further refactoring.
			t.Skip("checkMysqlConnection requires refactoring to be unit testable")
		})
	}
}

func TestCheckConnections(t *testing.T) {
	tests := []struct {
		name    string
		configs []types.MysqlConfig
		tries   int
		wantErr bool
	}{
		{
			name:    "empty config list returns no error",
			configs: []types.MysqlConfig{},
			tries:   3,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckConnections(tt.configs, tt.tries)

			if tt.wantErr {
				if err == nil {
					t.Error("CheckConnections() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("CheckConnections() unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
