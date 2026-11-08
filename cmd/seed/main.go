package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matheuseufrasio2/ponto_certo/internal/config"
	"github.com/matheuseufrasio2/ponto_certo/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	dbURL := os.Getenv("DATABASE_URL_UNPOOLED")
	if strings.TrimSpace(dbURL) == "" {
		dbURL = cfg.DatabaseURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pool, err := database.NewPostgresPool(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	seedPath := "scripts/seed.sql"
	if len(os.Args) > 1 {
		seedPath = os.Args[1]
	}

	content, err := os.ReadFile(seedPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read seed file %s: %v\n", seedPath, err)
		os.Exit(1)
	}

	fmt.Printf("🌱 Populando banco de dados com dados realistas (%s)...\n", seedPath)

	tx, err := pool.Begin(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to start transaction: %v\n", err)
		os.Exit(1)
	}

	if _, err := tx.Exec(ctx, string(content)); err != nil {
		_ = tx.Rollback(ctx)
		fmt.Fprintf(os.Stderr, "❌ Failed to execute seed script: %v\n", err)
		os.Exit(1)
	}

	if err := tx.Commit(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to commit seed transaction: %v\n", err)
		os.Exit(1)
	}

	printStats(ctx, pool)
}

func printStats(ctx context.Context, pool *pgxpool.Pool) {
	var orgCount, unitCount, userCount, barberCount, apptCount int
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM organizations").Scan(&orgCount)
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM barbershop_units").Scan(&unitCount)
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount)
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM barbers").Scan(&barberCount)
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM appointments").Scan(&apptCount)

	fmt.Println()
	fmt.Println("✅ Banco de dados populado com sucesso!")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("🏢 Organizações (Tenants): %d\n", orgCount)
	fmt.Printf("📍 Unidades Físicas:        %d\n", unitCount)
	fmt.Printf("👥 Usuários no Sistema:    %d\n", userCount)
	fmt.Printf("✂️  Barbeiros Operacionais:  %d\n", barberCount)
	fmt.Printf("📅 Agendamentos / Sessões:  %d\n", apptCount)
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("🔑 Senha padrão de todos os usuários: admin123")
	fmt.Println("👑 Super Admin: superadmin@pontocerto.com.br")
	fmt.Println()
}
