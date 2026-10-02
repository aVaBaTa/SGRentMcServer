# playrena-asa-server — ARK: Survival Ascended sous Proton, LISTÉ dans le browser.
#
# L'image sknnr (figée août 2024, GE-Proton 8-30) boote et se joint par IP, mais
# logue « Steam Subsystem initialized: FAILED » → le serveur n'est jamais annoncé
# aux clients Steam (invisible dans la liste non-officielle). Diagnostic 2026-10-02
# sur xe80dell, quatre correctifs nécessaires ensemble (chacun testé séparément) :
#   1. GE-Proton récent (10-34) — 8-30 est trop vieux pour le steamclient.so 2026.
#   2. Shims « legacycompat » : sans client Steam installé, Proton ne trouve aucun
#      steamclient64.dll à copier dans C:\Program Files (x86)\Steam\ → steam_api
#      échoue en 9 ms. Son propre lsteamclient.dll fait exactement ce rôle.
#   3. Suppression des DLL Steam Windows du dossier serveur avant le lancement
#      (steam_api les préfère et elles ne marchent pas sous Wine) — comme POK.
#   4. SteamAppId/SteamGameId en env : steam_api lit steam_appid.txt dans le cwd,
#      et l'entrypoint lance l'exe depuis /home/steam, pas depuis Win64.
# Résultat : « Steam Subsystem initialized: Success ».
FROM sknnr/ark-ascended-server:latest

ARG GE_PROTON=10-34
ENV GE_PROTON_VERSION=${GE_PROTON} \
    SteamAppId=2399830 \
    SteamGameId=2399830

USER steam
WORKDIR /home/steam

# 1. Proton récent (téléchargé une fois au build, ~520 Mo).
ADD --chown=steam:steam https://github.com/GloriousEggroll/proton-ge-custom/releases/download/GE-Proton${GE_PROTON}/GE-Proton${GE_PROTON}.tar.gz /tmp/ge.tar.gz
RUN set -e; cd "$STEAM_PATH/compatibilitytools.d" && tar xzf /tmp/ge.tar.gz && rm /tmp/ge.tar.gz \
    && test -x "GE-Proton${GE_PROTON}/proton" \
    # 2. Shims legacycompat = lsteamclient de ce Proton (64 + 32 bits).
    && P="$STEAM_PATH/compatibilitytools.d/GE-Proton${GE_PROTON}" \
    && mkdir -p "$STEAM_PATH/legacycompat" \
    && cp "$P/files/lib/wine/x86_64-windows/lsteamclient.dll" "$STEAM_PATH/legacycompat/steamclient64.dll" \
    && cp "$P/files/lib/wine/i386-windows/lsteamclient.dll" "$STEAM_PATH/legacycompat/steamclient.dll" \
    # 3. Nettoyage des DLL Steam Windows juste avant le `proton run` (après le
    #    steamcmd validate, qui les remettrait sinon).
    && sed -i 's|^\${STEAM_PATH}/compatibilitytools.d/GE-Proton|rm -f ${ARK_PATH}/ShooterGame/Binaries/Win64/steamclient*.dll ${ARK_PATH}/ShooterGame/Binaries/Win64/tier0_s*.dll ${ARK_PATH}/ShooterGame/Binaries/Win64/vstdlib_s*.dll\n&|' /home/steam/entrypoint.sh \
    && grep -q 'rm -f ${ARK_PATH}/ShooterGame/Binaries/Win64/steamclient' /home/steam/entrypoint.sh
