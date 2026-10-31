// Command server runs the Realmforge game API and the scheduled-event worker.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"realmforge/server/internal/api"
	"realmforge/server/internal/game"
	"realmforge/server/internal/pack"
	"realmforge/server/internal/platform"
	"realmforge/server/migrations"
)

func main() {
	cfg, err := platform.LoadConfig()
	log := platform.NewLogger(cfg.Env)
	if err != nil {
		log.Error("config", "err", err)
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := platform.OpenDB(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("database", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := platform.Migrate(ctx, db, migrations.FS, log); err != nil {
		log.Error("migrate", "err", err)
		os.Exit(1)
	}
	g := game.New(db, platform.SystemClock{}, log)

	// RF_WORLDS="slug:packDir[:Display Name[:speed]],..." creates missing worlds at startup.
	for _, spec := range strings.Split(os.Getenv("RF_WORLDS"), ",") {
		parts := strings.SplitN(strings.TrimSpace(spec), ":", 4)
		if len(parts) < 2 {
			continue
		}
		p, err := pack.LoadFile(filepath.Join(cfg.PacksDir, parts[1]))
		if err != nil {
			log.Error("pack invalid", "pack", parts[1], "err", err)
			os.Exit(1)
		}
		name, speed := parts[0], 1.0
		if len(parts) >= 3 && parts[2] != "" {
			name = parts[2]
		}
		if len(parts) == 4 {
			if speed, err = strconv.ParseFloat(parts[3], 64); err != nil {
				log.Error("world speed", "err", err)
				os.Exit(2)
			}
		}
		id, err := g.EnsureWorld(ctx, parts[0], name, p, time.Now().UnixNano()%1_000_000_007, speed)
		if err != nil {
			log.Error("world bootstrap", "err", err)
			os.Exit(1)
		}
		if cfg.AutoMigratePacks {
			// Worlds pin a pack snapshot; bring existing worlds up to the pack on
			// disk when every ID in use still exists (or is mapped by migrations).
			rep, err := g.MigrateWorldPack(ctx, 0, id, p, false)
			switch e, _ := game.AsError(err); {
			case err == nil && rep.Applied:
				log.Info("world pack migrated", "slug", parts[0], "from", rep.FromVersion, "to", rep.ToVersion, "rows", rep.Rows)
			case err == nil && len(rep.Missing) > 0:
				log.Warn("world pack NOT migrated: incompatible", "slug", parts[0], "missing", rep.Missing)
			case e != nil && e.Code == "same_pack":
			case err != nil:
				log.Warn("world pack migration skipped", "slug", parts[0], "err", err)
			}
		}
		log.Info("world ready", "slug", parts[0], "id", id, "pack", p.ID, "version", p.Version)
	}
	// RF_ADMIN_LOGIN promotes an existing login at startup. It is ignored in
	// production because anyone could register that login first (SECURITY S-2);
	// promote production admins explicitly in SQL instead.
	if cfg.AdminBootstrap != "" && cfg.Env == "production" {
		log.Warn("RF_ADMIN_LOGIN ignored in production; promote admins with SQL")
	} else if cfg.AdminBootstrap != "" {
		if _, err := db.Exec(ctx, `UPDATE accounts SET is_admin=true WHERE login=$1`, cfg.AdminBootstrap); err != nil {
			log.Warn("admin bootstrap", "err", err)
		}
	}
	if cfg.WorkerEnabled {
		go g.RunWorker(ctx, cfg.WorkerInterval)
	}
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: api.New(g, cfg, log),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 120 * time.Second}
	go func() {
		log.Info("listening", "addr", cfg.HTTPAddr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http", "err", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdown)
	log.Info("stopped")
}
