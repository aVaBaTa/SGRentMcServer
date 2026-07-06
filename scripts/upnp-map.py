#!/usr/bin/env python3
"""upnp-map.py — mapper des ports sur le routeur via UPnP IGD, sans root ni paquet.

Pourquoi : le NAT de la maison doit forwarder les ports de jeu vers xe80dell
(10.0.0.2). Pas de sudo → pas de miniupnpc ; ce script (stdlib seulement) fait
le SSDP + SOAP AddPortMapping, comme le module UPnP du serveur Calradia.

Usage :
  ./upnp-map.py status                          # gateway + IP externe
  ./upnp-map.py add 7777 tcp [--to 10.0.0.2]    # mappe un port
  ./upnp-map.py add 7777 both                   # TCP + UDP
  ./upnp-map.py add 25566-25581 both            # une plage (1 appel SOAP/port !)
  ./upnp-map.py del 7777 tcp
  ./upnp-map.py list                            # mappings existants

⚠️ La plupart des routeurs plafonnent le nombre d'entrées UPnP (~64-128) : ne
PAS mapper toute la plage Playrena (1000 ports) — pour ça, une règle manuelle
« port range forwarding » dans l'admin du routeur reste la bonne solution.
Bail par défaut 0 (permanent) ; certains routeurs refusent → réessaie avec
--lease 604800 (7 jours, à renouveler).
"""

import http.client
import re
import socket
import sys
import urllib.parse
import urllib.request
from xml.etree import ElementTree

# IP source des requêtes SOAP. Les box Helix/XB7 n'acceptent AddPortMapping
# QUE vers l'IP du demandeur → il faut émettre depuis l'IP cible du mapping
# (ex. 10.0.0.2 quand la machine a plusieurs adresses).
SOURCE_IP: str | None = None

SSDP_ADDR = ("239.255.255.250", 1900)
ST = "urn:schemas-upnp-org:device:InternetGatewayDevice:1"
WAN_SERVICES = (
    "urn:schemas-upnp-org:service:WANIPConnection:1",
    "urn:schemas-upnp-org:service:WANPPPConnection:1",
)


def local_ip() -> str:
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    s.connect(("8.8.8.8", 80))  # aucune donnée envoyée — juste pour choisir l'interface
    ip = s.getsockname()[0]
    s.close()
    return ip


def discover(timeout: float = 3.0) -> str:
    """SSDP M-SEARCH → URL de description du gateway."""
    msg = (
        "M-SEARCH * HTTP/1.1\r\n"
        f"HOST: {SSDP_ADDR[0]}:{SSDP_ADDR[1]}\r\n"
        'MAN: "ssdp:discover"\r\n'
        "MX: 2\r\n"
        f"ST: {ST}\r\n\r\n"
    ).encode()
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    s.settimeout(timeout)
    s.sendto(msg, SSDP_ADDR)
    try:
        while True:
            data, _ = s.recvfrom(4096)
            m = re.search(rb"(?im)^location:\s*(\S+)", data)
            if m:
                return m.group(1).decode()
    except socket.timeout:
        sys.exit("aucun gateway UPnP n'a répondu (UPnP désactivé sur le routeur ?)")
    finally:
        s.close()


def control_url(desc_url: str) -> tuple[str, str]:
    """Description XML → (URL de contrôle, type de service WAN*Connection)."""
    xml = urllib.request.urlopen(desc_url, timeout=5).read()
    root = ElementTree.fromstring(xml)
    ns = {"d": "urn:schemas-upnp-org:device-1-0"}
    base = re.match(r"(https?://[^/]+)", desc_url).group(1)
    for svc in root.iter("{urn:schemas-upnp-org:device-1-0}service"):
        stype = svc.findtext("d:serviceType", "", ns)
        if stype in WAN_SERVICES:
            path = svc.findtext("d:controlURL", "", ns)
            return (path if path.startswith("http") else base + path, stype)
    sys.exit("le gateway n'expose pas de service WAN(IP|PPP)Connection")


def soap(ctrl: str, stype: str, action: str, args: dict) -> str:
    body_args = "".join(f"<{k}>{v}</{k}>" for k, v in args.items())
    body = (
        '<?xml version="1.0"?>'
        '<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" '
        's:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">'
        f'<s:Body><u:{action} xmlns:u="{stype}">{body_args}</u:{action}></s:Body>'
        "</s:Envelope>"
    ).encode()
    u = urllib.parse.urlparse(ctrl)
    conn = http.client.HTTPConnection(
        u.hostname, u.port or 80, timeout=5,
        source_address=(SOURCE_IP, 0) if SOURCE_IP else None,
    )
    conn.request("POST", u.path, body=body, headers={
        "Content-Type": 'text/xml; charset="utf-8"',
        "SOAPAction": f'"{stype}#{action}"',
    })
    resp = conn.getresponse()
    detail = resp.read().decode(errors="replace")
    conn.close()
    if resp.status >= 400:
        code = re.search(r"<errorCode>(\d+)</errorCode>", detail)
        desc = re.search(r"<errorDescription>([^<]*)</errorDescription>", detail)
        sys.exit(
            f"{action} refusé par le routeur"
            + (f" : {code.group(1)} {desc.group(1) if desc else ''}" if code else f" (HTTP {resp.status})")
        )
    return detail


def parse_ports(spec: str) -> range:
    if "-" in spec:
        a, b = spec.split("-", 1)
        return range(int(a), int(b) + 1)
    return range(int(spec), int(spec) + 1)


def main() -> None:
    if len(sys.argv) < 2 or sys.argv[1] in ("-h", "--help"):
        sys.exit(__doc__)
    cmd = sys.argv[1]
    to_ip = local_ip()
    lease = "0"
    desc = None
    argv = []
    it = iter(sys.argv[2:])
    for a in it:
        if a == "--to":
            to_ip = next(it)
        elif a == "--lease":
            lease = next(it)
        elif a == "--desc":
            desc = next(it)
        else:
            argv.append(a)

    # Émettre depuis l'IP cible si c'est une IP de cette machine (exigence XB7).
    global SOURCE_IP
    try:
        probe = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        probe.bind((to_ip, 0))
        probe.close()
        SOURCE_IP = to_ip
    except OSError:
        pass  # cible = autre machine (ex. node2) → source par défaut

    # --desc court-circuite le SSDP : utile quand ufw bloque les réponses
    # multicast (cas xe80dell). Helix/Vidéotron (XB6/XB7) :
    #   --desc http://10.0.0.1:49152/IGDdevicedesc_brlan0.xml
    ctrl, stype = control_url(desc or discover())

    if cmd == "status":
        r = soap(ctrl, stype, "GetExternalIPAddress", {})
        ip = re.search(r"<NewExternalIPAddress>([^<]+)", r)
        print(f"gateway OK ({stype.rsplit(':', 2)[-2]}) — IP externe : {ip.group(1) if ip else '?'}")
        return

    if cmd == "list":
        i = 0
        while True:
            try:
                r = soap(ctrl, stype, "GetGenericPortMappingEntry", {"NewPortMappingIndex": i})
            except SystemExit:
                break
            port = re.search(r"<NewExternalPort>(\d+)", r)
            proto = re.search(r"<NewProtocol>(\w+)", r)
            client = re.search(r"<NewInternalClient>([^<]+)", r)
            desc = re.search(r"<NewPortMappingDescription>([^<]*)", r)
            print(f"{port.group(1):>5}/{proto.group(1).lower()} → {client.group(1)}  ({desc.group(1) if desc else ''})")
            i += 1
        return

    if cmd in ("add", "del"):
        if len(argv) < 2:
            sys.exit("usage : add|del <port|a-b> <tcp|udp|both> [--to IP] [--lease sec]")
        protos = ["TCP", "UDP"] if argv[1].lower() == "both" else [argv[1].upper()]
        for port in parse_ports(argv[0]):
            for proto in protos:
                if cmd == "add":
                    soap(ctrl, stype, "AddPortMapping", {
                        "NewRemoteHost": "",
                        "NewExternalPort": port,
                        "NewProtocol": proto,
                        "NewInternalPort": port,
                        "NewInternalClient": to_ip,
                        "NewEnabled": "1",
                        "NewPortMappingDescription": "playrena",
                        "NewLeaseDuration": lease,
                    })
                    print(f"mappé {port}/{proto.lower()} → {to_ip} (bail {lease}s)")
                else:
                    soap(ctrl, stype, "DeletePortMapping", {
                        "NewRemoteHost": "",
                        "NewExternalPort": port,
                        "NewProtocol": proto,
                    })
                    print(f"supprimé {port}/{proto.lower()}")
        return

    sys.exit(f"commande inconnue : {cmd}\n{__doc__}")


if __name__ == "__main__":
    main()
