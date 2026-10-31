package api

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"realmforge/server/internal/game"
	"realmforge/server/internal/pack"
)

func (s *Server) routes() {
	m := s.mux
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]any{"status": "ok"}) })
	m.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := s.G.DB.Ping(r.Context()); err != nil {
			writeJSON(w, 503, map[string]any{"status": "db_unavailable"})
			return
		}
		writeJSON(w, 200, map[string]any{"status": "ready", "time": s.G.Clock.Now()})
	})
	m.HandleFunc("GET /metrics", s.metrics)
	m.HandleFunc("GET /api/v1/time", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"data": map[string]any{"serverTime": s.G.Clock.Now()}})
	})

	// --- auth ---
	type creds struct {
		Login       string `json:"login"`
		Password    string `json:"password"`
		DisplayName string `json:"displayName,omitempty"`
	}
	m.HandleFunc("POST /api/v1/auth/register", s.ipLimited(func(w http.ResponseWriter, r *http.Request) {
		var c creds
		if err := decode(r, &c); err != nil {
			s.fail(w, r, err)
			return
		}
		sess, err := s.G.Register(r.Context(), c.Login, c.Password, c.DisplayName, s.Cfg.SessionTTL)
		s.respond(w, r, sess, err)
	}))
	m.HandleFunc("POST /api/v1/auth/login", s.ipLimited(func(w http.ResponseWriter, r *http.Request) {
		var c creds
		if err := decode(r, &c); err != nil {
			s.fail(w, r, err)
			return
		}
		sess, err := s.G.Login(r.Context(), c.Login, c.Password, s.Cfg.SessionTTL)
		s.respond(w, r, sess, err)
	}))
	m.HandleFunc("POST /api/v1/auth/guest", s.ipLimited(func(w http.ResponseWriter, r *http.Request) {
		sess, err := s.G.Guest(r.Context(), s.Cfg.SessionTTL)
		s.respond(w, r, sess, err)
	}))
	m.HandleFunc("POST /api/v1/auth/platform", s.ipLimited(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Platform   string `json:"platform"`
			ExternalID string `json:"externalId"`
			Signature  string `json:"signature,omitempty"`
		}
		if err := decode(r, &req); err != nil {
			s.fail(w, r, err)
			return
		}
		sess, err := s.G.PlatformLogin(r.Context(), req.Platform, req.ExternalID, s.Cfg.DevPlatformAuth, s.Cfg.SessionTTL)
		s.respond(w, r, sess, err)
	}))
	m.HandleFunc("POST /api/v1/auth/logout", s.auth(func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, map[string]bool{"ok": true}, s.G.Logout(r.Context(), bearer(r)))
	}))
	m.HandleFunc("POST /api/v1/auth/upgrade", s.auth(func(w http.ResponseWriter, r *http.Request) {
		var c creds
		if err := decode(r, &c); err != nil {
			s.fail(w, r, err)
			return
		}
		a, err := s.G.Upgrade(r.Context(), account(r).ID, c.Login, c.Password)
		s.respond(w, r, a, err)
	}))
	m.HandleFunc("GET /api/v1/me", s.auth(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"data": account(r)})
	}))
	m.HandleFunc("PUT /api/v1/me/locale", s.auth(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Locale string `json:"locale"`
		}
		if err := decode(r, &req); err != nil {
			s.fail(w, r, err)
			return
		}
		s.respond(w, r, map[string]bool{"ok": true}, s.G.SetLocale(r.Context(), account(r).ID, req.Locale))
	}))
	m.HandleFunc("DELETE /api/v1/me", s.auth(func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, map[string]bool{"deleted": true}, s.G.DeleteAccount(r.Context(), account(r).ID))
	}))

	// --- worlds ---
	m.HandleFunc("GET /api/v1/worlds", s.auth(func(w http.ResponseWriter, r *http.Request) {
		ws, err := s.G.ListWorlds(r.Context(), account(r).ID)
		s.respond(w, r, ws, err)
	}))
	m.HandleFunc("GET /api/v1/worlds/{wid}/pack", s.auth(func(w http.ResponseWriter, r *http.Request) {
		wid, err := pathID(r, "wid")
		if err != nil {
			s.fail(w, r, err)
			return
		}
		p, err := s.G.Pack(r.Context(), s.G.DB, wid)
		if err != nil {
			s.fail(w, r, err)
			return
		}
		etag := `"` + p.Hash + `"`
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", etag)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "private, max-age=0, must-revalidate")
		_, _ = w.Write([]byte(`{"data":`))
		_, _ = w.Write(p.JSON())
		_, _ = w.Write([]byte(`}`))
	}))
	m.HandleFunc("POST /api/v1/worlds/{wid}/join", s.auth(func(w http.ResponseWriter, r *http.Request) {
		wid, err := pathID(r, "wid")
		if err != nil {
			s.fail(w, r, err)
			return
		}
		var req struct {
			Name string `json:"name"`
			Race string `json:"race"`
		}
		if err := decode(r, &req); err != nil {
			s.fail(w, r, err)
			return
		}
		pid, err := s.G.Join(r.Context(), account(r).ID, wid, req.Name, req.Race)
		s.respond(w, r, map[string]int64{"playerId": pid}, err)
	}))
	// player-scoped reads under /worlds/{wid}
	s.playerRead("GET /api/v1/worlds/{wid}/overview", func(c game.PCtx, r *http.Request) (any, error) { return c.Overview() })
	s.playerRead("GET /api/v1/worlds/{wid}/quests", func(c game.PCtx, r *http.Request) (any, error) { return c.Quests() })
	s.playerRead("GET /api/v1/worlds/{wid}/missions", func(c game.PCtx, r *http.Request) (any, error) { return c.Missions() })
	s.playerRead("GET /api/v1/worlds/{wid}/heroes", func(c game.PCtx, r *http.Request) (any, error) { return c.Heroes() })
	s.playerRead("GET /api/v1/worlds/{wid}/territories", func(c game.PCtx, r *http.Request) (any, error) { return c.Territories() })
	s.playerRead("GET /api/v1/worlds/{wid}/reports", func(c game.PCtx, r *http.Request) (any, error) {
		return c.Reports(int64(queryInt(r, "before", 0)), queryInt(r, "limit", 30))
	})
	s.playerRead("GET /api/v1/worlds/{wid}/reports/{rid}", func(c game.PCtx, r *http.Request) (any, error) {
		id, err := pathID(r, "rid")
		if err != nil {
			return nil, err
		}
		return c.Report(id)
	})
	s.playerRead("GET /api/v1/worlds/{wid}/notifications", func(c game.PCtx, r *http.Request) (any, error) {
		items, unread, err := c.Notifications(int64(queryInt(r, "after", 0)), queryInt(r, "limit", 50))
		return map[string]any{"items": items, "unread": unread}, err
	})
	s.playerRead("GET /api/v1/worlds/{wid}/messages", func(c game.PCtx, r *http.Request) (any, error) {
		return c.Messages(r.URL.Query().Get("box"), int64(queryInt(r, "before", 0)), queryInt(r, "limit", 50))
	})
	s.playerRead("GET /api/v1/worlds/{wid}/alliance-invites", func(c game.PCtx, r *http.Request) (any, error) { return c.Invites() })

	// world-scoped public reads (membership required)
	s.worldRead("GET /api/v1/worlds/{wid}/map", func(tx pgx.Tx, wid int64, r *http.Request) (any, error) {
		return s.G.Map(r.Context(), tx, wid, queryInt(r, "x", 0), queryInt(r, "y", 0), queryInt(r, "w", 16), queryInt(r, "h", 16), s.G.Clock.Now())
	})
	s.worldRead("GET /api/v1/worlds/{wid}/rankings", func(tx pgx.Tx, wid int64, r *http.Request) (any, error) {
		return s.G.Rankings(r.Context(), tx, wid, min(100, max(1, queryInt(r, "limit", 50))), max(0, queryInt(r, "offset", 0)))
	})
	s.worldRead("GET /api/v1/worlds/{wid}/market", func(tx pgx.Tx, wid int64, r *http.Request) (any, error) {
		var mine int64
		if r.URL.Query().Get("mine") == "1" {
			pid, err := s.G.PlayerID(r.Context(), tx, account(r).ID, wid)
			if err != nil {
				return nil, err
			}
			mine = pid
		}
		return s.G.Orders(r.Context(), tx, wid, r.URL.Query().Get("sell"), r.URL.Query().Get("buy"), mine)
	})
	s.worldRead("GET /api/v1/worlds/{wid}/alliances", func(tx pgx.Tx, wid int64, r *http.Request) (any, error) {
		return s.G.Alliances(r.Context(), tx, wid)
	})
	s.worldRead("GET /api/v1/worlds/{wid}/alliances/{aid}", func(tx pgx.Tx, wid int64, r *http.Request) (any, error) {
		aid, err := pathID(r, "aid")
		if err != nil {
			return nil, err
		}
		return s.G.Alliance(r.Context(), tx, wid, aid)
	})

	// --- player-scoped commands ---
	type idReq struct {
		ID int64 `json:"id"`
	}
	playerCmd(s, "POST /api/v1/worlds/{wid}/missions/{mid}/recall", func(c game.PCtx, r *http.Request, _ struct{}) (any, error) {
		id, err := pathID(r, "mid")
		if err != nil {
			return nil, err
		}
		return c.Recall(id)
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/quests/{qid}/claim", func(c game.PCtx, r *http.Request, req struct {
		CityID int64 `json:"cityId"`
	}) (any, error) {
		rewards, err := c.ClaimQuest(r.PathValue("qid"), req.CityID)
		return map[string]any{"rewards": rewards}, err
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/heroes/{hid}/skills", func(c game.PCtx, r *http.Request, req struct {
		Skill string `json:"skill"`
	}) (any, error) {
		id, err := pathID(r, "hid")
		if err != nil {
			return nil, err
		}
		return c.LearnSkill(id, req.Skill)
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/heroes/{hid}/equip", func(c game.PCtx, r *http.Request, req struct {
		ItemID int64 `json:"itemId"`
	}) (any, error) {
		id, err := pathID(r, "hid")
		if err != nil {
			return nil, err
		}
		return ok(c.Equip(id, req.ItemID))
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/heroes/{hid}/unequip", func(c game.PCtx, r *http.Request, req struct {
		Slot string `json:"slot"`
	}) (any, error) {
		id, err := pathID(r, "hid")
		if err != nil {
			return nil, err
		}
		return ok(c.Unequip(id, req.Slot))
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/market/{oid}/cancel", func(c game.PCtx, r *http.Request, _ struct{}) (any, error) {
		id, err := pathID(r, "oid")
		if err != nil {
			return nil, err
		}
		return c.CancelOrder(id)
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/territories/abandon", func(c game.PCtx, r *http.Request, req struct {
		X int `json:"x"`
		Y int `json:"y"`
	}) (any, error) {
		return ok(c.Abandon(req.X, req.Y))
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/alliances", func(c game.PCtx, r *http.Request, req struct {
		Name        string `json:"name"`
		Tag         string `json:"tag"`
		Description string `json:"description"`
	}) (any, error) {
		return c.CreateAlliance(req.Name, req.Tag, req.Description)
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/alliance/invites", func(c game.PCtx, r *http.Request, req struct {
		Player string `json:"player"`
	}) (any, error) {
		return c.Invite(req.Player)
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/alliance-invites/{iid}/respond", func(c game.PCtx, r *http.Request, req struct {
		Accept bool `json:"accept"`
	}) (any, error) {
		id, err := pathID(r, "iid")
		if err != nil {
			return nil, err
		}
		return ok(c.RespondInvite(id, req.Accept))
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/alliance/leave", func(c game.PCtx, r *http.Request, _ struct{}) (any, error) {
		return ok(c.Leave())
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/alliance/members/{pid}/role", func(c game.PCtx, r *http.Request, req struct {
		Role string `json:"role"`
	}) (any, error) {
		id, err := pathID(r, "pid")
		if err != nil {
			return nil, err
		}
		return ok(c.SetRole(id, req.Role))
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/alliance/relations", func(c game.PCtx, r *http.Request, req struct {
		AllianceID int64  `json:"allianceId"`
		Status     string `json:"status"`
	}) (any, error) {
		return ok(c.SetRelation(req.AllianceID, req.Status))
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/messages", func(c game.PCtx, r *http.Request, req struct {
		To       string `json:"to"`
		Alliance bool   `json:"alliance"`
		Subject  string `json:"subject"`
		Body     string `json:"body"`
	}) (any, error) {
		id, err := c.SendMessage(req.To, req.Alliance, req.Subject, req.Body)
		return map[string]int64{"id": id}, err
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/messages/{id}/read", func(c game.PCtx, r *http.Request, _ struct{}) (any, error) {
		id, err := pathID(r, "id")
		if err != nil {
			return nil, err
		}
		return ok(c.MarkMessageRead(id))
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/messages/{id}/report", func(c game.PCtx, r *http.Request, _ struct{}) (any, error) {
		id, err := pathID(r, "id")
		if err != nil {
			return nil, err
		}
		return ok(c.ReportMessage(id))
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/blocks", func(c game.PCtx, r *http.Request, req struct {
		PlayerID int64 `json:"playerId"`
		Block    bool  `json:"block"`
	}) (any, error) {
		return ok(c.Block(req.PlayerID, req.Block))
	})
	playerCmd(s, "POST /api/v1/worlds/{wid}/notifications/read", func(c game.PCtx, r *http.Request, req struct {
		UpTo int64 `json:"upTo"`
	}) (any, error) {
		return ok(c.MarkNotificationsRead(req.UpTo))
	})
	_ = idReq{}

	// --- city-scoped ---
	m.HandleFunc("GET /api/v1/cities/{cid}", s.auth(func(w http.ResponseWriter, r *http.Request) {
		cid, err := pathID(r, "cid")
		if err != nil {
			s.fail(w, r, err)
			return
		}
		s.read(w, r, func(tx pgx.Tx) (any, error) {
			return s.G.WithCity(r.Context(), tx, account(r).ID, cid, func(c game.PCtx, ci *game.City) (any, error) { return c.View(ci) })
		})
	}))
	cityCmd(s, "POST /api/v1/cities/{cid}/build", func(c game.PCtx, ci *game.City, req struct {
		Building string `json:"building"`
	}) (any, error) {
		return c.Build(ci, req.Building)
	})
	cityCmd(s, "POST /api/v1/cities/{cid}/research", func(c game.PCtx, ci *game.City, req struct {
		Tech string `json:"tech"`
	}) (any, error) {
		return c.Research(ci, req.Tech)
	})
	cityCmd(s, "POST /api/v1/cities/{cid}/train", func(c game.PCtx, ci *game.City, req struct {
		Unit     string `json:"unit"`
		Quantity int64  `json:"quantity"`
	}) (any, error) {
		return c.Train(ci, req.Unit, req.Quantity)
	})
	cityCmd(s, "POST /api/v1/cities/{cid}/jobs/cancel", func(c game.PCtx, ci *game.City, req struct {
		JobID int64 `json:"jobId"`
	}) (any, error) {
		refund, err := c.CancelJob(ci, req.JobID)
		return map[string]any{"refund": refund}, err
	})
	cityCmd(s, "POST /api/v1/cities/{cid}/rename", func(c game.PCtx, ci *game.City, req struct {
		Name string `json:"name"`
	}) (any, error) {
		return ok(c.Rename(ci, req.Name))
	})
	cityCmd(s, "POST /api/v1/cities/{cid}/missions", func(c game.PCtx, ci *game.City, req game.MissionRequest) (any, error) {
		return c.SendMission(ci, req)
	})
	cityCmd(s, "POST /api/v1/cities/{cid}/heroes", func(c game.PCtx, ci *game.City, req struct {
		Class string `json:"class"`
		Name  string `json:"name"`
	}) (any, error) {
		return c.HireHero(ci, req.Class, req.Name)
	})
	cityCmd(s, "POST /api/v1/cities/{cid}/market/orders", func(c game.PCtx, ci *game.City, req struct {
		Sell       string `json:"sell"`
		SellAmount int64  `json:"sellAmount"`
		Buy        string `json:"buy"`
		BuyAmount  int64  `json:"buyAmount"`
	}) (any, error) {
		return c.PlaceOrder(ci, req.Sell, req.SellAmount, req.Buy, req.BuyAmount)
	})
	cityCmd(s, "POST /api/v1/cities/{cid}/market/accept", func(c game.PCtx, ci *game.City, req struct {
		OrderID int64 `json:"orderId"`
	}) (any, error) {
		return c.AcceptOrder(ci, req.OrderID)
	})
	cityCmd(s, "POST /api/v1/cities/{cid}/exchange", func(c game.PCtx, ci *game.City, req struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount int64  `json:"amount"`
	}) (any, error) {
		return c.Exchange(ci, req.From, req.To, req.Amount)
	})
	cityCmd(s, "POST /api/v1/cities/{cid}/collect", func(c game.PCtx, ci *game.City, req struct {
		X int `json:"x"`
		Y int `json:"y"`
	}) (any, error) {
		return c.Collect(ci, req.X, req.Y)
	})

	// --- admin ---
	m.HandleFunc("GET /api/v1/admin/players/{pid}", s.admin(func(w http.ResponseWriter, r *http.Request) {
		pid, err := pathID(r, "pid")
		if err != nil {
			s.fail(w, r, err)
			return
		}
		out, err := s.G.AdminInspect(r.Context(), account(r).ID, pid)
		s.respond(w, r, out, err)
	}))
	m.HandleFunc("GET /api/v1/admin/events/failed", s.admin(func(w http.ResponseWriter, r *http.Request) {
		out, err := s.G.FailedEvents(r.Context())
		s.respond(w, r, out, err)
	}))
	m.HandleFunc("POST /api/v1/admin/events/{eid}/retry", s.admin(func(w http.ResponseWriter, r *http.Request) {
		eid, err := pathID(r, "eid")
		if err != nil {
			s.fail(w, r, err)
			return
		}
		s.respond(w, r, map[string]bool{"ok": true}, s.G.RetryEvent(r.Context(), account(r).ID, eid))
	}))
	m.HandleFunc("POST /api/v1/admin/cities/{cid}/grant", s.admin(func(w http.ResponseWriter, r *http.Request) {
		cid, err := pathID(r, "cid")
		if err != nil {
			s.fail(w, r, err)
			return
		}
		var req struct {
			Amounts pack.Amounts `json:"amounts"`
			Reason  string       `json:"reason"`
		}
		if err := decode(r, &req); err != nil {
			s.fail(w, r, err)
			return
		}
		s.respond(w, r, map[string]bool{"ok": true}, s.G.AdminGrant(r.Context(), account(r).ID, cid, req.Amounts, req.Reason))
	}))
	m.HandleFunc("POST /api/v1/admin/accounts/{aid}/ban", s.admin(func(w http.ResponseWriter, r *http.Request) {
		aid, err := pathID(r, "aid")
		if err != nil {
			s.fail(w, r, err)
			return
		}
		var req struct {
			Until  time.Time `json:"until"`
			Reason string    `json:"reason"`
		}
		if err := decode(r, &req); err != nil {
			s.fail(w, r, err)
			return
		}
		s.respond(w, r, map[string]bool{"ok": true}, s.G.Ban(r.Context(), account(r).ID, aid, req.Until, req.Reason))
	}))
	m.HandleFunc("GET /api/v1/admin/audit", s.admin(func(w http.ResponseWriter, r *http.Request) {
		out, err := s.G.Audit(r.Context(), min(500, max(1, queryInt(r, "limit", 100))))
		s.respond(w, r, out, err)
	}))
	m.HandleFunc("GET /api/v1/admin/messages/reported", s.admin(func(w http.ResponseWriter, r *http.Request) {
		out, err := s.G.ReportedMessages(r.Context())
		s.respond(w, r, out, err)
	}))
	m.HandleFunc("POST /api/v1/admin/worlds/{wid}/migrate-pack", s.admin(func(w http.ResponseWriter, r *http.Request) {
		wid, err := pathID(r, "wid")
		if err != nil {
			s.fail(w, r, err)
			return
		}
		var req struct {
			PackDir string `json:"packDir"` // directory name under RF_PACKS_DIR
			DryRun  bool   `json:"dryRun"`
		}
		if err := decode(r, &req); err != nil {
			s.fail(w, r, err)
			return
		}
		if strings.ContainsAny(req.PackDir, "/\\.") || req.PackDir == "" {
			s.fail(w, r, &game.Error{Status: 400, Code: "invalid_request", Message: "packDir must be a plain directory name"})
			return
		}
		p, err := pack.LoadFile(filepath.Join(s.Cfg.PacksDir, req.PackDir))
		if err != nil {
			s.fail(w, r, &game.Error{Status: 422, Code: "invalid_pack", Message: err.Error()})
			return
		}
		rep, err := s.G.MigrateWorldPack(r.Context(), account(r).ID, wid, p, req.DryRun)
		s.respond(w, r, rep, err)
	}))

	// pack assets (icons, art) — served read-only from RF_PACKS_DIR/<pack>/assets
	m.HandleFunc("GET /packs/{pack}/assets/{file...}", func(w http.ResponseWriter, r *http.Request) {
		name, file := r.PathValue("pack"), r.PathValue("file")
		if strings.Contains(name, "..") || strings.Contains(file, "..") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=3600")
		http.ServeFile(w, r, filepath.Join(s.Cfg.PacksDir, name, "assets", filepath.Clean("/"+file)))
	})
	// optional: serve the built web client (SPA fallback)
	if s.Cfg.ClientDir != "" {
		fs := http.FileServer(http.Dir(s.Cfg.ClientDir))
		m.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			p := filepath.Join(s.Cfg.ClientDir, filepath.Clean("/"+r.URL.Path))
			if st, err := os.Stat(p); err != nil || st.IsDir() {
				f, err := os.Open(filepath.Join(s.Cfg.ClientDir, "index.html"))
				if err != nil {
					http.NotFound(w, r)
					return
				}
				defer f.Close()
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = io.Copy(w, f)
				return
			}
			fs.ServeHTTP(w, r)
		})
	}
}

func ok(err error) (any, error) {
	if err != nil {
		return nil, err
	}
	return map[string]bool{"ok": true}, nil
}

func (s *Server) respond(w http.ResponseWriter, r *http.Request, v any, err error) {
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]any{"data": v})
}

func (s *Server) playerRead(pattern string, fn func(c game.PCtx, r *http.Request) (any, error)) {
	s.mux.HandleFunc(pattern, s.auth(func(w http.ResponseWriter, r *http.Request) {
		wid, err := pathID(r, "wid")
		if err != nil {
			s.fail(w, r, err)
			return
		}
		s.read(w, r, func(tx pgx.Tx) (any, error) {
			return s.G.WithPlayer(r.Context(), tx, account(r).ID, wid, func(c game.PCtx) (any, error) { return fn(c, r) })
		})
	}))
}

func (s *Server) worldRead(pattern string, fn func(tx pgx.Tx, wid int64, r *http.Request) (any, error)) {
	s.mux.HandleFunc(pattern, s.auth(func(w http.ResponseWriter, r *http.Request) {
		wid, err := pathID(r, "wid")
		if err != nil {
			s.fail(w, r, err)
			return
		}
		s.read(w, r, func(tx pgx.Tx) (any, error) {
			if _, err := s.G.PlayerID(r.Context(), tx, account(r).ID, wid); err != nil {
				return nil, err
			}
			return fn(tx, wid, r)
		})
	}))
}

func playerCmd[T any](s *Server, pattern string, fn func(c game.PCtx, r *http.Request, req T) (any, error)) {
	s.mux.HandleFunc(pattern, s.auth(func(w http.ResponseWriter, r *http.Request) {
		wid, err := pathID(r, "wid")
		if err != nil {
			s.fail(w, r, err)
			return
		}
		cmd(s, w, r, func(tx pgx.Tx, req T) (any, error) {
			return s.G.WithPlayer(r.Context(), tx, account(r).ID, wid, func(c game.PCtx) (any, error) { return fn(c, r, req) })
		})
	}))
}

func cityCmd[T any](s *Server, pattern string, fn func(c game.PCtx, ci *game.City, req T) (any, error)) {
	s.mux.HandleFunc(pattern, s.auth(func(w http.ResponseWriter, r *http.Request) {
		cid, err := pathID(r, "cid")
		if err != nil {
			s.fail(w, r, err)
			return
		}
		cmd(s, w, r, func(tx pgx.Tx, req T) (any, error) {
			return s.G.WithCity(r.Context(), tx, account(r).ID, cid, func(c game.PCtx, ci *game.City) (any, error) { return fn(c, ci, req) })
		})
	}))
}
