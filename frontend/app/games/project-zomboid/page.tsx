"use client";

import { Biohazard } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { GameHostingPage } from "@/components/game-hosting-page";

export default function ZomboidHosting() {
  const { t } = useI18n();
  return <GameHostingPage gameId="project-zomboid" icon={Biohazard} texts={t.gamePages["project-zomboid"]} />;
}
