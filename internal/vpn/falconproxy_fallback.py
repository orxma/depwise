#!/usr/bin/env python3
"""Falcon Proxy (safe local fallback).

Open, self-contained HTTP/HTTPS (CONNECT) proxy used by the Privanox bot
when the upstream FirewallFalcon binary is unavailable.

This is a clean, auditable implementation: it only forwards connection
requests to their stated destinations. It does NOT phone home, modify
/etc/hosts, install root CAs, or alter system trust stores.

Usage:
    falconproxy -p <port> [other flags ignored]
"""
import socket
import sys
import threading

BUFFER = 65536


def parse_port(argv):
    port = 8080
    i = 0
    while i < len(argv):
        a = argv[i]
        if a in ("-p", "--port") and i + 1 < len(argv):
            try:
                port = int(argv[i + 1])
            except ValueError:
                pass
            i += 2
            continue
        if a.startswith("-p") and len(a) > 2:
            try:
                port = int(a[2:])
            except ValueError:
                pass
            i += 1
            continue
        # bare numeric argument is also accepted as the port
        if a.isdigit():
            try:
                port = int(a)
            except ValueError:
                pass
        i += 1
    return port


def relay(src, dst):
    try:
        while True:
            data = src.recv(BUFFER)
            if not data:
                break
            dst.sendall(data)
    except Exception:
        pass
    finally:
        for s in (src, dst):
            try:
                s.shutdown(socket.SHUT_RDWR)
            except Exception:
                pass
            try:
                s.close()
            except Exception:
                pass


def handle_client(client):
    try:
        client.settimeout(10)
        request = client.recv(BUFFER)
        if not request:
            client.close()
            return
        client.settimeout(None)

        first_line = request.split(b"\r\n", 1)[0].decode("utf-8", "ignore")
        parts = first_line.split()
        if len(parts) < 3:
            client.close()
            return
        method, target, _ = parts[0], parts[1], parts[2]

        if method.upper() == "CONNECT":
            host, _, port = target.partition(":")
            port = int(__import__("re").sub(r"\D", "", port)) if port else 443
            try:
                remote = socket.create_connection((host, port), timeout=15)
            except Exception:
                client.sendall(b"HTTP/1.1 502 Bad Gateway\r\n\r\n")
                client.close()
                return
            client.sendall(b"HTTP/1.1 200 Connection Established\r\n\r\n")
            threading.Thread(target=relay, args=(client, remote), daemon=True).start()
            threading.Thread(target=relay, args=(remote, client), daemon=True).start()
            return

        host = port = None
        if "://" in target:
            rest = target.split("://", 1)[1]
            host = rest.split("/", 1)[0]
        else:
            host = target
        if ":" in host:
            host, port = host.split(":", 1)
            port = int(port)
        else:
            port = 80

        try:
            remote = socket.create_connection((host, port), timeout=15)
        except Exception:
            client.sendall(b"HTTP/1.1 502 Bad Gateway\r\n\r\n")
            client.close()
            return

        remote.sendall(request)
        threading.Thread(target=relay, args=(client, remote), daemon=True).start()
        threading.Thread(target=relay, args=(remote, client), daemon=True).start()
    except Exception:
        try:
            client.close()
        except Exception:
            pass


def main():
    port = parse_port(sys.argv[1:])
    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server.bind(("0.0.0.0", port))
    server.listen(128)
    while True:
        try:
            conn, _ = server.accept()
        except Exception:
            continue
        threading.Thread(target=handle_client, args=(conn,), daemon=True).start()


if __name__ == "__main__":
    main()
