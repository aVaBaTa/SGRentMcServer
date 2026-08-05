"use client";

import { Skull } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { GameHostingPage } from "@/components/game-hosting-page";

export default function SdtdHosting() {
  const { t } = useI18n();
  return <GameHostingPage gameId="7dtd" icon={Skull} texts={t.gamePages["7dtd"]} />;
}
