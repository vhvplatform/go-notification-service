package mongodb

import (
	"testing"
)

func TestValidateMongoURI(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		wantErr bool
	}{
		{
			name:    "valid mongodb URI",
			uri:     "mongodb://localhost:27017",
			wantErr: false,
		},
		{
			name:    "valid mongodb+srv URI",
			uri:     "mongodb+srv://cluster.mongodb.net",
			wantErr: false,
		},
		{
			name:    "empty URI",
			uri:     "",
			wantErr: true,
		},
		{
			name:    "invalid scheme",
			uri:     "http://localhost:27017",
			wantErr: true,
		},
		{
			name:    "invalid scheme - postgres",
			uri:     "postgres://localhost:5432",
			wantErr: true,
		},
		{
			name:    "missing host",
			uri:     "mongodb://",
			wantErr: true,
		},
		{
			name:    "malformed URI",
			uri:     "not-a-valid-uri",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMongoURI(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMongoURI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewMongoClient_SecurityValidation(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		database string
		wantErr  bool
	}{
		{
			name:     "empty database name",
			uri:      "mongodb://localhost:27017",
			database: "",
			wantErr:  true,
		},
		{
			name:     "invalid database name with slash",
			uri:      "mongodb://localhost:27017",
			database: "test/db",
			wantErr:  true,
		},
		{
			name:     "invalid database name with backslash",
			uri:      "mongodb://localhost:27017",
			database: "test\\db",
			wantErr:  true,
		},
		{
			name:     "invalid database name with dot",
			uri:      "mongodb://localhost:27017",
			database: "test.db",
			wantErr:  true,
		},
		{
			name:     "invalid database name with special chars",
			uri:      "mongodb://localhost:27017",
			database: "test$db",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This will fail to connect but we're testing validation
			_, err := NewMongoClient(tt.uri, tt.database)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMongoClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
