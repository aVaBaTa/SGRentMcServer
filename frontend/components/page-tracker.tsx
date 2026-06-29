"use client";

import { useEffect, useRef } from "react";
import { usePathname } from "next/navigation";

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
const HEARTBEAT_MS = 20000; // battement de présence tant que l'onglet est ouvert

// visitorId : identifiant stable par navigateur (localStorage) → « 1 personne = 1 vid ».
// Plusieurs onglets du même navigateur partagent le vid (comptés une fois).
function visitorId(): string {
  try {
    let id = localStorage.getItem("pv_vid");
    if (!id) {
      id = crypto.randomUUID?.() ?? String(Math.random()).slice(2);
      localStorage.setItem("pv_vid", id);
    }
    return id;
  } catch {
    return "anon";
  }
}

// PageTracker : (1) envoie une vue de page à chaque changement de route (analytics /admin) ;
// (2) envoie un battement de présence régulier pour le compteur « en ligne maintenant ».
// credentials inclus → le backend attache le pseudo si le visiteur est connecté (cookie JWT).
export function PageTracker() {
  const pathname = usePathname();
  const last = useRef<string | null>(null);
  const pathRef = useRef(pathname);
  pathRef.current = pathname;

  // (1) Vue de page au changement de route.
  useEffect(() => {
    if (!pathname || pathname === last.current) return;
    const previous = last.current;
    last.current = pathname;
    fetch(`${API}/api/v1/track`, {
      method: "POST",
      credentials: "include",
      keepalive: true,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ path: pathname, referer: previous }),
    }).catch(() => {});
  }, [pathname]);

  // (2) Heartbeat de présence (toutes les HEARTBEAT_MS + au retour sur l'onglet).
  useEffect(() => {
    const vid = visitorId();
    const beat = () =>
      fetch(`${API}/api/v1/presence`, {
        method: "POST",
        credentials: "include",
        keepalive: true,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ vid, path: pathRef.current }),
      }).catch(() => {});
    beat();
    const iv = setInterval(beat, HEARTBEAT_MS);
    const onVis = () => {
      if (document.visibilityState === "visible") beat();
    };
    document.addEventListener("visibilitychange", onVis);
    return () => {
      clearInterval(iv);
      document.removeEventListener("visibilitychange", onVis);
    };
  }, []);

  return null;
}
