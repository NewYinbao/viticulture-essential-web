"""Windows native smoke; never uses live saves or launcher port 3012."""
import json
import os
from pathlib import Path
import socket
import subprocess
import tempfile
import time
import urllib.request
import urllib.error


def main():
    if os.name != "nt":
        raise SystemExit("Run with native Windows Python, not WSL Python")
    root = Path(__file__).resolve().parent
    exe = root / "Viticulture-EE-Rules-Fix.exe"
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        port = sock.getsockname()[1]
    base = "http://127.0.0.1:" + str(port)
    http = urllib.request.build_opener(urllib.request.ProxyHandler({}))
    def api(path, body=None, token=None):
        headers = {"Content-Type": "application/json"}
        if token:
            headers["Authorization"] = "Bearer " + token
        req = urllib.request.Request(base + path,
            data=None if body is None else json.dumps(body).encode(), headers=headers)
        with http.open(req, timeout=3) as response:
            return json.load(response)
    with tempfile.TemporaryDirectory(prefix="viticulture-native-smoke-") as td:
        proc = None
        def start():
            nonlocal proc
            proc = subprocess.Popen([str(exe), "-addr", "127.0.0.1:" + str(port),
                                     "-data", str(Path(td) / "data")],
                                    cwd=str(root), stdout=subprocess.DEVNULL,
                                    stderr=subprocess.DEVNULL)
            deadline = time.monotonic() + 15
            while time.monotonic() < deadline:
                if proc.poll() is not None:
                    raise RuntimeError("Server exited: " + str(proc.returncode))
                try:
                    health = api("/api/health")
                    assert health == {"ok": True, "version": "ee-rules-fix"}, health
                    return
                except (OSError, urllib.error.URLError):
                    time.sleep(0.1)
            raise TimeoutError("Health check timed out")
        def stop():
            nonlocal proc
            if proc is not None:
                if proc.poll() is None:
                    proc.terminate()
                proc.wait(timeout=10)
                proc = None
        try:
            start()
            with http.open(base + "/", timeout=3) as response:
                assert response.status == 200
                assert b"<html" in response.read().lower()
            try:
                api("/api/state")
                raise AssertionError("Unauthenticated state was accepted")
            except urllib.error.HTTPError as err:
                assert err.code == 401
            host = api("/api/create", {"name": "Native Smoke A"})
            guest = api("/api/join", {"name": "Native Smoke B", "code": host["code"]})
            assert guest["code"] == host["code"]
            state = api("/api/state", token=host["token"])
            assert state["phase"] == "lobby" and len(state["players"]) == 2
            state = api("/api/action", {"type": "start", "revision": state["revision"]}, host["token"])
            assert state["phase"] != "lobby"
            assert (Path(td) / "data" / "ee-state-v1.json").is_file()
            stop()
            start()
            restored = api("/api/state", token=host["token"])
            assert restored["revision"] == state["revision"]
            assert restored["phase"] == state["phase"]
            print(json.dumps({"completed": True, "nativeWindows": True,
                              "healthVersion": "ee-rules-fix", "port": port,
                              "cases": ["health", "embedded UI", "authentication", "create", "join", "start", "persist/restart"],
                              "liveSavesTouched": False}))
        finally:
            stop()


if __name__ == "__main__":
    main()
