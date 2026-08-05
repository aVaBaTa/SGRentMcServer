"use client";

import { PawPrint } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { GameHostingPage } from "@/components/game-hosting-page";

export default function PalworldHosting() {
  const { t } = useI18n();
  return <GameHostingPage gameId="palworld" icon={PawPrint} texts={t.gamePages.palworld} />;
}
