package server

import "testing"

// Logs RÉELS capturés au boot d'un container ghcr.io/terkea/hytale-server (2026-06-11,
// validation #G). Le downloader (step 1 OAuth) émet une URL « complète » avec user_code,
// une URL nue, une ligne « Authorization code: », et une URL de MAJ (.zip) à ignorer.
const realHytaleDownloaderLog = `
========================================
  Hytale Server Docker Container
========================================

[INFO] Setting up user with UID=1000 and GID=1000
[SUCCESS] User setup complete
[INFO] Downloading server files using Hytale Downloader CLI...

========================================
  HYTALE DOWNLOADER
========================================

A new version of hytale-downloader is available: 2026.05.13-99ade04 (current: 2026.01.09-49e5904)Download it from: https://downloader.hytale.com/hytale-downloader.zip ***

Please visit the following URL to authenticate:
https://oauth.accounts.hytale.com/oauth2/device/verify?user_code=rcv7cG7L
Or visit the following URL and enter the code:
https://oauth.accounts.hytale.com/oauth2/device/verify
Authorization code: rcv7cG7L
`

// Step 2 : authentification du serveur (device flow du entrypoint.sh).
const realHytaleServerAuthLog = `
========================================
  SERVER AUTHENTICATION REQUIRED
========================================

  Visit: https://oauth.accounts.hytale.com/oauth2/device/verify?user_code=ABCD-9999
  Code:  ABCD-9999

  Waiting for authorization...
`

// Ligne émise par l'entrypoint une fois les tokens en place, juste avant exec java.
const realHytaleReadyLog = `
[SUCCESS] Server authenticated and ready!
========================================
  Hytale Server Starting
========================================
`

// Step 1 : OAuth downloader. URL complète (1 clic) préférée, .zip ignoré.
func TestParseRealHytaleStep1(t *testing.T) {
	st, done := parseAuthFromLogs(realHytaleDownloaderLog)
	if done {
		t.Fatal("ne devrait pas être 'done' pendant l'OAuth")
	}
	if !st.Pending {
		t.Fatal("devrait être pending")
	}
	if st.Code != "rcv7cG7L" {
		t.Fatalf("code = %q, attendu rcv7cG7L", st.Code)
	}
	// L'URL retournée doit être la version COMPLÈTE (avec user_code), pas la nue ni le .zip.
	want := "https://oauth.accounts.hytale.com/oauth2/device/verify?user_code=rcv7cG7L"
	if st.URL != want {
		t.Fatalf("url = %q, attendu %q", st.URL, want)
	}
	if st.Step != 1 {
		t.Fatalf("step = %d, attendu 1", st.Step)
	}
}

// Step 2 : authentification serveur.
func TestParseRealHytaleStep2(t *testing.T) {
	st, done := parseAuthFromLogs(realHytaleServerAuthLog)
	if done {
		t.Fatal("ne devrait pas être 'done'")
	}
	if st.Code != "ABCD-9999" {
		t.Fatalf("code = %q, attendu ABCD-9999", st.Code)
	}
	if st.Step != 2 {
		t.Fatalf("step = %d, attendu 2", st.Step)
	}
}

// Serveur prêt → done=true (le watcher passe en 'running').
func TestParseRealHytaleReady(t *testing.T) {
	_, done := parseAuthFromLogs(realHytaleReadyLog)
	if !done {
		t.Fatal("la ligne 'authenticated and ready' devrait marquer done=true")
	}
}
