package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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

	// Droit "création illimitée" : seul un user qui a ce droit peut créer
	// directement un serveur sur un plan PAYANT sans paiement. Sinon le plan
	// demandé est ramené à "free" (le passage payant se fait ensuite via le
	// checkout PayPal qui upgrade le serveur).
	allowPaid := false
	if u, err := s.userRepo.GetByID(r.Context(), claims.UserID); err == nil {
		allowPaid = u.UnlimitedCreate
	}

	gs, status, err := s.createServerForUser(r.Context(), claims.UserID, claims.Username, req, allowPaid)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	respond(w, http.StatusCreated, gs)
}

// createServerForUser = cœur partagé de la création de serveur (utilisé par le
// handler utilisateur ET par l'admin). allowPaid=false force un plan payant vers
// "free". Retourne le serveur, un code HTTP (en cas d'erreur) et l'erreur.
func (s *Server) createServerForUser(ctx context.Context, userID, username string, req createServerRequest, allowPaid bool) (*servers.GameServer, int, error) {
	if req.Name == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("name is required")
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

	gameDef, err := servers.GetGame(req.Game)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	plan, err := servers.GetPlan(req.Plan)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	// Garde paiement : un plan payant non autorisé est ramené à "free".
	if !plan.Free && !allowPaid {
		plan, _ = servers.GetPlan("free")
		req.Plan = plan.Name
	}

	// Plancher de ressources par jeu : un serveur Satisfactory tourne en 4 Go
	// même sur le plan "free" (offert pour l'instant). Le label de plan et la
	// facturation restent inchangés ; seules les ressources réelles sont relevées.
	ramMb, cpuCores := gameDef.ApplyFloor(plan.RAMMb, plan.CPUCores)

	node, err := s.orch.BestNode(ctx, ramMb)
	if err != nil {
		slog.Error("no available node", "err", err)
		return nil, http.StatusServiceUnavailable, fmt.Errorf("no capacity available, try again later")
	}

	port, err := servers.NextAvailablePort(ctx, s.db)
	if err != nil {
		slog.Error("no available port", "err", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("no available port")
	}

	subdomain := s.uniqueSubdomain(ctx, username, req.Name)

	gs := &servers.GameServer{
		UserID:    userID,
		Name:      req.Name,
		Subdomain: subdomain,
		Game:      req.Game,
		Plan:      req.Plan,
		Version:   req.Version,
		Node:      node.ID,
		Status:    "creating",
		RAMMb:     ramMb,
		CPUCores:  cpuCores,
		Port:      port,
	}

	if err := s.serverRepo.Create(ctx, gs); err != nil {
		slog.Error("failed to save server", "err", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("database error")
	}

	// La création du container (pull image + start) peut prendre plusieurs
	// minutes. On provisionne en arrière-plan ; le dashboard poll le statut.
	go s.provisionServer(gs, req.Game, port, subdomain, plan)
	return gs, http.StatusCreated, nil
}

// buildSpec assemble le ServerSpec à partir de la définition du jeu. Les
// ressources (RAM/CPU) proviennent du GameServer (déjà passées au plancher du
// jeu) ; les ports, le volume, l'env et le routing dépendent du jeu.
func (s *Server) buildSpec(gs *servers.GameServer, gameDef servers.GameDef, plan servers.Plan, version string) orchestrator.ServerSpec {
	spec := orchestrator.ServerSpec{
		ContainerName: fmt.Sprintf("sgrent-%s", gs.ID),
		Image:         gameDef.Image,
		RAMMb:         gs.RAMMb,
		CPUCores:      gs.CPUCores,
		DataPath:      gameDef.DataPath,
		Subdomain:     gs.Subdomain,
		Network:       s.cfg.MCNetwork,
		EnvVars:       gameDef.Env(plan, version, gs.Port),
	}
	for _, p := range gameDef.Ports(gs.Port) {
		spec.Ports = append(spec.Ports, orchestrator.PortBinding{
			HostPort: p.HostPort, Internal: p.Internal, Proto: p.Proto,
		})
	}
	// Routing par hostname réservé à Minecraft (mc-router). Les autres jeux
	// passent par l'accès direct IP:port.
	if gameDef.UsesMCRouter {
		spec.RouterHost = gs.Subdomain + "." + s.cfg.ServersDomain
		spec.RouterPort = gameDef.RouterPort
	}
	// Si les fichiers de jeu pré-téléchargés sont disponibles, on désactive le
	// téléchargement (et son OAuth) dans le container : le volume sera seedé avant
	// le start (cf. provisionServer / seedServer). Dernière occurrence = celle qui
	// gagne côté Docker, donc ces overrides priment sur l'env du jeu.
	if s.seedReady(gameDef) {
		spec.EnvVars = append(spec.EnvVars, "AUTO_DOWNLOAD=false", "SKIP_DOWNLOAD=true")
	}
	return spec
}

// seedDir retourne le dossier des fichiers pré-téléchargés d'un jeu, ou "" si non configuré.
func (s *Server) seedDir(gameID string) string {
	if s.cfg.SeedDir == "" {
		return ""
	}
	return filepath.Join(s.cfg.SeedDir, gameID)
}

// seedReady : tous les SeedFiles du jeu existent sous <SeedDir>/<game>/. Si oui, on
// seede le volume au lieu de laisser le container télécharger (Hytale).
func (s *Server) seedReady(gameDef servers.GameDef) bool {
	dir := s.seedDir(gameDef.ID)
	if dir == "" || len(gameDef.SeedFiles) == 0 {
		return false
	}
	for _, f := range gameDef.SeedFiles {
		if st, err := os.Stat(filepath.Join(dir, f)); err != nil || st.IsDir() || st.Size() == 0 {
			return false
		}
	}
	return true
}

// seedServer copie les fichiers de jeu pré-téléchargés dans le volume du container
// (créé mais pas encore démarré). À appeler entre CreateServer et StartServer.
func (s *Server) seedServer(ctx context.Context, node *orchestrator.Node, containerID string, gameDef servers.GameDef) error {
	dir := s.seedDir(gameDef.ID)
	for _, f := range gameDef.SeedFiles {
		if err := node.CopyFileToContainer(ctx, containerID, filepath.Join(dir, f), gameDef.DataPath); err != nil {
			return err
		}
	}
	return nil
}

// provisionServer crée et démarre le container en arrière-plan.
func (s *Server) provisionServer(gs *servers.GameServer, game string, port int, subdomain string, plan servers.Plan) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	gameDef, err := servers.GetGame(game)
	if err != nil {
		slog.Error("provision: unknown game", "id", gs.ID, "game", game, "err", err)
		s.serverRepo.UpdateStatus(ctx, gs.ID, "error")
		return
	}

	node, err := s.orch.NodeByID(gs.Node)
	if err != nil {
		slog.Error("provision: node unavailable", "id", gs.ID, "err", err)
		s.serverRepo.UpdateStatus(ctx, gs.ID, "error")
		return
	}

	containerID, err := node.CreateServer(ctx, s.buildSpec(gs, gameDef, plan, gs.Version))
	if err != nil {
		slog.Error("provision: create container failed", "id", gs.ID, "err", err)
		s.serverRepo.UpdateStatus(ctx, gs.ID, "error")
		return
	}
	s.serverRepo.UpdateContainerID(ctx, gs.ID, containerID)
	gs.ContainerID = containerID

	// Seed des fichiers de jeu pré-téléchargés dans le volume AVANT le start (Hytale) :
	// le container saute le téléchargement (et son OAuth downloader). Le client n'aura
	// que l'auth SERVEUR à faire. Si le seed échoue, on stoppe (AUTO_DOWNLOAD=false →
	// le container ne pourrait pas récupérer les fichiers seul).
	if s.seedReady(gameDef) {
		if err := s.seedServer(ctx, node, containerID, gameDef); err != nil {
			slog.Error("provision: seed failed", "id", gs.ID, "err", err)
			s.serverRepo.UpdateStatus(ctx, gs.ID, "error")
			return
		}
		slog.Info("provision: volume seedé avec fichiers pré-téléchargés", "id", gs.ID, "game", game)
	}

	if err := node.StartServer(ctx, containerID); err != nil {
		slog.Error("provision: start container failed", "id", gs.ID, "err", err)
		s.serverRepo.UpdateStatus(ctx, gs.ID, "error")
		return
	}

	// Jeux à authentification interactive (Hytale) : le serveur affiche un OAuth
	// device-code au 1er démarrage. On délègue à un watcher détaché (contexte
	// long) qui passe le statut à "auth_required" puis "running" une fois l'auth
	// complétée. Les autres jeux passent directement à "running".
	if gameDef.NeedsAuth {
		go s.watchAuthAndRun(gs, node)
		slog.Info("provision: awaiting interactive auth", "id", gs.ID, "port", port)
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

	gameDef, err := servers.GetGame(gs.Game)
	if err != nil {
		slog.Error("recreate: unknown game", "id", gs.ID, "game", gs.Game, "err", err)
		s.serverRepo.UpdateStatus(ctx, gs.ID, "error")
		return
	}

	if gs.ContainerID != "" {
		node.StopServer(ctx, gs.ContainerID)
		if err := node.RemoveServer(ctx, gs.ContainerID); err != nil {
			slog.Error("recreate: remove old container failed", "id", gs.ID, "err", err)
		}
	}

	containerID, err := node.CreateServer(ctx, s.buildSpec(gs, gameDef, plan, version))
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

	// Plancher de ressources par jeu (même logique qu'à la création) : un
	// downgrade ne doit pas descendre un serveur Satisfactory sous son minimum.
	gameDef, err := servers.GetGame(gs.Game)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ramMb, cpuCores := gameDef.ApplyFloor(plan.RAMMb, plan.CPUCores)

	// NOTE: la facturation (PayPal) est gérée séparément (handlers_billing).
	// Le nombre de slots (MAX_PLAYERS) est une variable d'env immuable : pour
	// qu'il évolue avec le plan, on recrée le container (RAM, CPU et MAX_PLAYERS
	// appliqués d'un coup). Le volume de données — donc le monde — est conservé.
	if err := s.serverRepo.UpdatePlan(r.Context(), gs.ID, plan.Name, ramMb, cpuCores); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	gs.Plan = plan.Name
	gs.RAMMb = ramMb
	gs.CPUCores = cpuCores

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

// registerRoute enregistre la route mc-router : hostname → IP_LAN_du_node:port.
// Fonctionne pour n'importe quel node (routing cross-host). Réservé aux jeux qui
// utilisent mc-router (Minecraft) ; les autres passent par l'accès IP:port direct.
func (s *Server) registerRoute(ctx context.Context, gs *servers.GameServer) {
	if def, err := servers.GetGame(gs.Game); err != nil || !def.UsesMCRouter {
		return
	}
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
	if def, err := servers.GetGame(gs.Game); err != nil || !def.UsesMCRouter {
		return
	}
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
