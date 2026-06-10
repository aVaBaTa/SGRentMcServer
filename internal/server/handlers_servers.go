package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/aVaBaTa/SGRentMcServer/internal/auth"
	"github.com/aVaBaTa/SGRentMcServer/internal/orchestrator"
	"github.com/aVaBaTa/SGRentMcServer/internal/servers"
	"github.com/go-chi/chi/v5"
)

type createServerRequest struct {
	Name    string `json:"name"`
	Plan    string `json:"plan"`
	Game    string `json:"game"`
	Version string `json:"version"`
}

func (s *Server) handleCreateServer(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)

	var req createServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.Game == "" {
		req.Game = "minecraft"
	}
	if req.Plan == "" {
		req.Plan = "free"
	}
	if req.Version == "" {
		req.Version = "LATEST"
	}

	plan, err := servers.GetPlan(req.Plan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Choisir le meilleur node
	node, err := s.orch.BestNode(r.Context(), plan.RAMMb)
	if err != nil {
		slog.Error("no available node", "err", err)
		http.Error(w, "no capacity available, try again later", http.StatusServiceUnavailable)
		return
	}

	// Port disponible
	port, err := servers.NextAvailablePort(r.Context(), s.db)
	if err != nil {
		slog.Error("no available port", "err", err)
		http.Error(w, "no available port", http.StatusInternalServerError)
		return
	}

	subdomain := s.uniqueSubdomain(r.Context(), claims.Username, req.Name)

	gs := &servers.GameServer{
		UserID:    claims.UserID,
		Name:      req.Name,
		Subdomain: subdomain,
		Game:      req.Game,
		Plan:      req.Plan,
		Version:   req.Version,
		Node:      node.ID,
		Status:    "creating",
		RAMMb:     plan.RAMMb,
		CPUCores:  plan.CPUCores,
		Port:      port,
	}

	if err := s.serverRepo.Create(r.Context(), gs); err != nil {
		slog.Error("failed to save server", "err", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// La création du container (pull image + start) peut prendre plusieurs
	// minutes. On répond immédiatement (statut "creating") et on provisionne
	// en arrière-plan ; le dashboard poll le statut.
	go s.provisionServer(gs, req.Game, port, subdomain, plan)

	respond(w, http.StatusCreated, gs)
}

// provisionServer crée et démarre le container en arrière-plan.
func (s *Server) provisionServer(gs *servers.GameServer, game string, port int, subdomain string, plan servers.Plan) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	node, err := s.orch.NodeByID(gs.Node)
	if err != nil {
		slog.Error("provision: node unavailable", "id", gs.ID, "err", err)
		s.serverRepo.UpdateStatus(ctx, gs.ID, "error")
		return
	}

	containerID, err := node.CreateServer(ctx, orchestrator.ServerSpec{
		ContainerName: fmt.Sprintf("sgrent-%s", gs.ID),
		Image:         gameImage(game),
		RAMMb:         plan.RAMMb,
		CPUCores:      plan.CPUCores,
		Port:          port,
		Subdomain:     subdomain,
		RouterHost:    subdomain + "." + s.cfg.ServersDomain,
		Network:       s.cfg.MCNetwork,
		EnvVars:       minecraftEnv(plan, gs.Version),
	})
	if err != nil {
		slog.Error("provision: create container failed", "id", gs.ID, "err", err)
		s.serverRepo.UpdateStatus(ctx, gs.ID, "error")
		return
	}
	s.serverRepo.UpdateContainerID(ctx, gs.ID, containerID)

	if err := node.StartServer(ctx, containerID); err != nil {
		slog.Error("provision: start container failed", "id", gs.ID, "err", err)
		s.serverRepo.UpdateStatus(ctx, gs.ID, "error")
		return
	}

	s.serverRepo.UpdateStatus(ctx, gs.ID, "running")
	s.registerRoute(ctx, gs)
	slog.Info("provision: server running", "id", gs.ID, "port", port)
}

func (s *Server) handleListServers(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)

	list, err := s.serverRepo.ListByUser(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []*servers.GameServer{}
	}
	respond(w, http.StatusOK, list)
}

func (s *Server) handleGetServer(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)
	id := chi.URLParam(r, "id")

	gs, err := s.serverRepo.GetByID(r.Context(), id, claims.UserID)
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}
	respond(w, http.StatusOK, gs)
}

func (s *Server) handleDeleteServer(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)
	id := chi.URLParam(r, "id")

	gs, err := s.serverRepo.GetByID(r.Context(), id, claims.UserID)
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	s.unregisterRoute(r.Context(), gs)

	if gs.ContainerID != "" {
		node, err := s.orch.NodeByID(gs.Node)
		if err == nil {
			node.RemoveServer(r.Context(), gs.ContainerID)
		}
	}

	if err := s.serverRepo.Delete(r.Context(), id, claims.UserID); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusOK, map[string]string{"msg": "server deleted"})
}

func (s *Server) handleStartServer(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := node.StartServer(r.Context(), gs.ContainerID); err != nil {
		http.Error(w, "failed to start server", http.StatusInternalServerError)
		return
	}
	s.serverRepo.UpdateStatus(r.Context(), gs.ID, "running")
	respond(w, http.StatusOK, map[string]string{"status": "running"})
}

func (s *Server) handleStopServer(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := node.StopServer(r.Context(), gs.ContainerID); err != nil {
		http.Error(w, "failed to stop server", http.StatusInternalServerError)
		return
	}
	s.serverRepo.UpdateStatus(r.Context(), gs.ID, "stopped")
	respond(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (s *Server) handleRestartServer(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := node.RestartServer(r.Context(), gs.ContainerID); err != nil {
		http.Error(w, "failed to restart server", http.StatusInternalServerError)
		return
	}
	s.serverRepo.UpdateStatus(r.Context(), gs.ID, "running")
	respond(w, http.StatusOK, map[string]string{"status": "running"})
}

type versionRequest struct {
	Version string `json:"version"`
}

func (s *Server) handleChangeVersion(w http.ResponseWriter, r *http.Request) {
	var req versionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Version == "" {
		http.Error(w, "version is required", http.StatusBadRequest)
		return
	}

	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	plan, err := servers.GetPlan(gs.Plan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Les variables d'env d'un container sont immuables → on recrée le container
	// (le volume de données, donc le monde, est conservé).
	s.serverRepo.UpdateVersion(r.Context(), gs.ID, req.Version)
	s.serverRepo.UpdateStatus(r.Context(), gs.ID, "creating")
	go s.recreateServer(gs, node, plan, req.Version)

	gs.Version = req.Version
	gs.Status = "creating"
	respond(w, http.StatusOK, gs)
}

func (s *Server) recreateServer(gs *servers.GameServer, node *orchestrator.Node, plan servers.Plan, version string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if gs.ContainerID != "" {
		node.StopServer(ctx, gs.ContainerID)
		if err := node.RemoveServer(ctx, gs.ContainerID); err != nil {
			slog.Error("recreate: remove old container failed", "id", gs.ID, "err", err)
		}
	}

	containerID, err := node.CreateServer(ctx, orchestrator.ServerSpec{
		ContainerName: fmt.Sprintf("sgrent-%s", gs.ID),
		Image:         gameImage(gs.Game),
		RAMMb:         plan.RAMMb,
		CPUCores:      plan.CPUCores,
		Port:          gs.Port,
		Subdomain:     gs.Subdomain,
		RouterHost:    gs.Subdomain + "." + s.cfg.ServersDomain,
		Network:       s.cfg.MCNetwork,
		EnvVars:       minecraftEnv(plan, version),
	})
	if err != nil {
		slog.Error("recreate: create container failed", "id", gs.ID, "err", err)
		s.serverRepo.UpdateStatus(ctx, gs.ID, "error")
		return
	}
	s.serverRepo.UpdateContainerID(ctx, gs.ID, containerID)

	if err := node.StartServer(ctx, containerID); err != nil {
		slog.Error("recreate: start failed", "id", gs.ID, "err", err)
		s.serverRepo.UpdateStatus(ctx, gs.ID, "error")
		return
	}
	s.serverRepo.UpdateStatus(ctx, gs.ID, "running")
	s.registerRoute(ctx, gs)
	slog.Info("recreate: server running new version", "id", gs.ID, "version", version)
}

type upgradeRequest struct {
	Plan string `json:"plan"`
}

func (s *Server) handleUpgradeServer(w http.ResponseWriter, r *http.Request) {
	var req upgradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	plan, err := servers.GetPlan(req.Plan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// NOTE: la facturation (PayPal) est gérée séparément (handlers_billing).
	// Le nombre de slots (MAX_PLAYERS) est une variable d'env immuable : pour
	// qu'il évolue avec le plan, on recrée le container (RAM, CPU et MAX_PLAYERS
	// appliqués d'un coup). Le volume de données — donc le monde — est conservé.
	if err := s.serverRepo.UpdatePlan(r.Context(), gs.ID, plan.Name, plan.RAMMb, plan.CPUCores); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	gs.Plan = plan.Name
	gs.RAMMb = plan.RAMMb
	gs.CPUCores = plan.CPUCores

	if gs.ContainerID != "" {
		s.serverRepo.UpdateStatus(r.Context(), gs.ID, "creating")
		go s.recreateServer(gs, node, plan, gs.Version)
		gs.Status = "creating"
	}
	respond(w, http.StatusOK, gs)
}

// resolveServer récupère le GameServer et son Node depuis l'URL.
func (s *Server) resolveServer(r *http.Request) (*servers.GameServer, *orchestrator.Node, error) {
	claims := auth.GetClaims(r)
	id := chi.URLParam(r, "id")

	gs, err := s.serverRepo.GetByID(r.Context(), id, claims.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("server not found")
	}
	node, err := s.orch.NodeByID(gs.Node)
	if err != nil {
		return nil, nil, fmt.Errorf("node unavailable")
	}
	return gs, node, nil
}

func gameImage(game string) string {
	switch game {
	case "minecraft":
		return "itzg/minecraft-server:latest"
	default:
		return "itzg/minecraft-server:latest"
	}
}

// minecraftEnv : config pour l'image itzg/minecraft-server.
// TYPE=PAPER → performant + support plugins. Moddable via Fabric/Forge plus tard.
// version : "LATEST", "1.21.4", etc.
func minecraftEnv(plan servers.Plan, version string) []string {
	if version == "" {
		version = "LATEST"
	}
	// MEMORY = tas JVM = la RAM annoncée du plan. La limite mémoire du container
	// (HostConfig.Memory) est volontairement plus haute pour laisser de la marge
	// au non-heap (metaspace, threads, buffers directs, GC) — voir container.go.
	env := []string{
		"EULA=TRUE",
		"TYPE=PAPER",
		"VERSION=" + version,
		fmt.Sprintf("MEMORY=%dM", plan.RAMMb),
		"USE_AIKAR_FLAGS=true",
	}
	if plan.MaxSlots > 0 {
		env = append(env, fmt.Sprintf("MAX_PLAYERS=%d", plan.MaxSlots))
	}
	return env
}

// registerRoute enregistre la route mc-router : hostname → IP_LAN_du_node:port.
// Fonctionne pour n'importe quel node (routing cross-host).
func (s *Server) registerRoute(ctx context.Context, gs *servers.GameServer) {
	if !s.mcRouter.Configured() {
		return
	}
	host := gs.Subdomain + "." + s.cfg.ServersDomain

	// Node local : mc-router joint le container par nom sur mc-net (DNS Docker).
	// Node distant : via l'IP LAN + le port hôte publié.
	var backend string
	if node, err := s.orch.NodeByID(gs.Node); err == nil && node.IsLocal() {
		backend = fmt.Sprintf("sgrent-%s:25565", gs.ID)
	} else {
		addr := s.cfg.NodeAddrs[gs.Node]
		if addr == "" {
			slog.Warn("no LAN address for node", "node", gs.Node)
			return
		}
		backend = fmt.Sprintf("%s:%d", addr, gs.Port)
	}

	if err := s.mcRouter.Register(ctx, host, backend); err != nil {
		slog.Warn("mc-router register failed", "host", host, "backend", backend, "err", err)
	} else {
		slog.Info("route registered", "host", host, "backend", backend)
	}
}

func (s *Server) unregisterRoute(ctx context.Context, gs *servers.GameServer) {
	if s.mcRouter.Configured() {
		s.mcRouter.Unregister(ctx, gs.Subdomain+"."+s.cfg.ServersDomain)
	}
}

// ReconcileRoutes ré-enregistre les routes de tous les serveurs (au démarrage de l'API).
func (s *Server) ReconcileRoutes(ctx context.Context) {
	list, err := s.serverRepo.ListAll(ctx)
	if err != nil {
		slog.Error("reconcile routes: list failed", "err", err)
		return
	}
	for _, gs := range list {
		if gs.Status == "running" || gs.Status == "creating" {
			s.registerRoute(ctx, gs)
		}
	}
	slog.Info("routes reconciled", "count", len(list))
}

// uniqueSubdomain génère un sous-domaine unique basé sur le username.
// 1er serveur → "username" ; suivants → "username-<nom>" puis "username-<n>".
func (s *Server) uniqueSubdomain(ctx context.Context, username, name string) string {
	base := sanitizeSubdomain(username)
	if base == "" {
		base = "srv"
	}

	candidates := []string{base}
	if ns := sanitizeSubdomain(name); ns != "" {
		candidates = append(candidates, base+"-"+ns)
	}
	for _, c := range candidates {
		if exists, err := s.serverRepo.SubdomainExists(ctx, c); err == nil && !exists {
			return c
		}
	}
	for i := 2; i < 1000; i++ {
		c := fmt.Sprintf("%s-%d", base, i)
		if exists, err := s.serverRepo.SubdomainExists(ctx, c); err == nil && !exists {
			return c
		}
	}
	return fmt.Sprintf("%s-%d", base, time.Now().UnixNano()%100000)
}

func sanitizeSubdomain(username string) string {
	s := strings.ToLower(username)
	var b strings.Builder
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			b.WriteRune(c)
		}
	}
	return b.String()
}
