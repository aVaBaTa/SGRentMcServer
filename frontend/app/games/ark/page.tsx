"use client";

import { Bone } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { GameHostingPage } from "@/components/game-hosting-page";

export default function ArkHosting() {
  const { t } = useI18n();
  return <GameHostingPage gameId="ark" icon={Bone} texts={t.gamePages.ark} />;
}
