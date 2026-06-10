import { redirect } from "next/navigation";

// La home redirige vers la page Minecraft (cf. aussi next.config.ts).
export default function Home() {
  redirect("/games/minecraft");
}
