package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matheuseufrasio2/ponto_certo/internal/config"
	"github.com/matheuseufrasio2/ponto_certo/internal/database"
)

type migrationFile struct {
	version int
	name    string
	upPath  string
	downPath string
}

func main() {
	cmd := "up"
	if len(os.Args) > 1 {
		cmd = strings.ToLower(strings.TrimSpace(os.Args[1]))
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Prefer unpooled URL for DDL migrations if provided (Neon requirement for transaction DDL)
	dbURL := os.Getenv("DATABASE_URL_UNPOOLED")
	if strings.TrimSpace(dbURL) == "" {
		dbURL = cfg.DatabaseURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := database.NewPostgresPool(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := run(ctx, pool, cmd); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Migration command failed: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, pool *pgxpool.Pool, cmd string) error {
	// 1. Ensure schema_migrations table exists (compatible with golang-migrate)
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT NOT NULL PRIMARY KEY,
			dirty BOOLEAN NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// 2. Discover migrations from disk
	migrations, err := loadMigrations("migrations")
	if err != nil {
		return fmt.Errorf("load migrations from disk: %w", err)
	}

	// 3. Resolve current version (and self-heal if schema_migrations is uninitialized but tables exist)
	currentVersion, dirty, err := getCurrentVersion(ctx, pool)
	if err != nil {
		return fmt.Errorf("resolve current version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in a dirty state at version %d (manual recovery required)", currentVersion)
	}

	switch cmd {
	case "status":
		return printStatus(migrations, currentVersion)

	case "up":
		return applyUp(ctx, pool, migrations, currentVersion)

	case "down":
		return applyDown(ctx, pool, migrations, currentVersion)

	default:
		return fmt.Errorf("unknown migration command %q. Valid commands: up, down, status", cmd)
	}
}

func getCurrentVersion(ctx context.Context, pool *pgxpool.Pool) (int64, bool, error) {
	var version int64
	var dirty bool

	err := pool.QueryRow(ctx, "SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version, &dirty)
	if err == nil {
		return version, dirty, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, false, err
	}

	// Fallback/auto-detect if schema_migrations was empty but tables were previously created
	detectedVersion := detectExistingSchemaVersion(ctx, pool)
	if detectedVersion > 0 {
		slog.Info("🔍 Detected existing database schema without schema_migrations entry, synchronizing baseline", "baseline_version", detectedVersion)
		_, err = pool.Exec(ctx, "INSERT INTO schema_migrations (version, dirty) VALUES ($1, FALSE)", detectedVersion)
		if err != nil {
			return 0, false, fmt.Errorf("record detected baseline migration version: %w", err)
		}
		return detectedVersion, false, nil
	}

	return 0, false, nil
}

func detectExistingSchemaVersion(ctx context.Context, pool *pgxpool.Pool) int64 {
	// Check if past_due_since exists in organizations (v6)
	var colExists bool
	_ = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'organizations' AND column_name = 'past_due_since'
		);
	`).Scan(&colExists)
	if colExists {
		return 6
	}

	// Check if subscription_status exists in organizations (v5)
	_ = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'organizations' AND column_name = 'subscription_status'
		);
	`).Scan(&colExists)
	if colExists {
		return 5
	}

	// Check if unit_operating_hours exists (v4)
	var tblExists bool
	_ = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'unit_operating_hours'
		);
	`).Scan(&tblExists)
	if tblExists {
		return 4
	}

	// Check if appointments table exists (v3)
	_ = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'appointments'
		);
	`).Scan(&tblExists)
	if tblExists {
		return 3
	}

	// Check if users table exists (v2)
	_ = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'users'
		);
	`).Scan(&tblExists)
	if tblExists {
		return 2
	}

	// Check if barbers table exists (v1)
	_ = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = 'barbers'
		);
	`).Scan(&tblExists)
	if tblExists {
		return 1
	}

	return 0
}

func loadMigrations(dir string) ([]migrationFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	fileMap := make(map[int]*migrationFile)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		parts := strings.Split(name, "_")
		if len(parts) < 2 {
			continue
		}

		v, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		m, exists := fileMap[v]
		if !exists {
			m = &migrationFile{
				version: v,
				name:    strings.TrimSuffix(strings.TrimSuffix(name, ".up.sql"), ".down.sql"),
			}
			fileMap[v] = m
		}

		if strings.HasSuffix(name, ".up.sql") {
			m.upPath = filepath.Join(dir, name)
		} else if strings.HasSuffix(name, ".down.sql") {
			m.downPath = filepath.Join(dir, name)
		}
	}

	migrations := make([]migrationFile, 0, len(fileMap))
	for _, m := range fileMap {
		migrations = append(migrations, *m)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

func printStatus(migrations []migrationFile, currentVersion int64) error {
	fmt.Println()
	fmt.Println("📊 Status das Migrations do Banco de Dados:")
	fmt.Println(strings.Repeat("-", 75))
	fmt.Printf("%-8s | %-45s | %-15s\n", "VERSÃO", "MIGRATION", "STATUS")
	fmt.Println(strings.Repeat("-", 75))

	for _, m := range migrations {
		status := "⏳ PENDENTE"
		if int64(m.version) <= currentVersion {
			status = "✅ APLICADA"
		}
		displayName := m.name
		if len(displayName) > 45 {
			displayName = displayName[:42] + "..."
		}
		fmt.Printf("%06d   | %-45s | %-15s\n", m.version, displayName, status)
	}
	fmt.Println(strings.Repeat("-", 75))
	fmt.Printf("Versão atual do banco: %06d\n\n", currentVersion)
	return nil
}

func applyUp(ctx context.Context, pool *pgxpool.Pool, migrations []migrationFile, currentVersion int64) error {
	applied := 0

	for _, m := range migrations {
		if int64(m.version) <= currentVersion {
			continue
		}

		if m.upPath == "" {
			return fmt.Errorf("migration %06d missing .up.sql file", m.version)
		}

		content, err := os.ReadFile(m.upPath)
		if err != nil {
			return fmt.Errorf("read migration file %s: %w", m.upPath, err)
		}

		fmt.Printf("🚀 Aplicando migration %06d: %s...\n", m.version, m.name)

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin transaction: %w", err)
		}

		// Mark dirty = TRUE in case statement crashes mid-flight
		_, _ = tx.Exec(ctx, "DELETE FROM schema_migrations")
		_, _ = tx.Exec(ctx, "INSERT INTO schema_migrations (version, dirty) VALUES ($1, TRUE)", m.version)

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("execute SQL for migration %06d (%s): %w", m.version, m.name, err)
		}

		// Mark dirty = FALSE
		_, err = tx.Exec(ctx, "UPDATE schema_migrations SET dirty = FALSE WHERE version = $1", m.version)
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("finalize migration version record: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %06d: %w", m.version, err)
		}

		currentVersion = int64(m.version)
		applied++
		fmt.Printf("✅ Migration %06d aplicada com sucesso!\n", m.version)
	}

	if applied == 0 {
		fmt.Printf("✨ O banco de dados já está totalmente atualizado (versão %06d).\n", currentVersion)
	} else {
		fmt.Printf("🎉 %d migration(s) aplicada(s) com sucesso. Versão atual: %06d.\n", applied, currentVersion)
	}

	return nil
}

func applyDown(ctx context.Context, pool *pgxpool.Pool, migrations []migrationFile, currentVersion int64) error {
	if currentVersion == 0 {
		fmt.Println("ℹ️ O banco de dados já está na versão inicial 0. Nenhuma migration a reverter.")
		return nil
	}

	var target *migrationFile
	var prevVersion int64

	for i, m := range migrations {
		if int64(m.version) == currentVersion {
			target = &migrations[i]
			if i > 0 {
				prevVersion = int64(migrations[i-1].version)
			}
			break
		}
	}

	if target == nil || target.downPath == "" {
		return fmt.Errorf("down migration for version %06d not found or missing file", currentVersion)
	}

	content, err := os.ReadFile(target.downPath)
	if err != nil {
		return fmt.Errorf("read down migration file %s: %w", target.downPath, err)
	}

	fmt.Printf("⏪ Revertendo migration %06d: %s...\n", target.version, target.name)

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	_, _ = tx.Exec(ctx, "UPDATE schema_migrations SET dirty = TRUE WHERE version = $1", currentVersion)

	if _, err := tx.Exec(ctx, string(content)); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("execute down SQL for migration %06d: %w", currentVersion, err)
	}

	_, _ = tx.Exec(ctx, "DELETE FROM schema_migrations")
	if prevVersion > 0 {
		_, err = tx.Exec(ctx, "INSERT INTO schema_migrations (version, dirty) VALUES ($1, FALSE)", prevVersion)
	}
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("update schema_migrations to version %06d: %w", prevVersion, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit down migration: %w", err)
	}

	fmt.Printf("✅ Migration %06d revertida com sucesso. Versão atual: %06d.\n", currentVersion, prevVersion)
	return nil
}
