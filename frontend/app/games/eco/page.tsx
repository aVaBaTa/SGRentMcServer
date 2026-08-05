"use client";

import { Leaf } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { GameHostingPage } from "@/components/game-hosting-page";

export default function EcoHosting() {
  const { t } = useI18n();
  return <GameHostingPage gameId="eco" icon={Leaf} texts={t.gamePages.eco} />;
}
