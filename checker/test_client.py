import json
import threading
import time
import unittest
from collections import deque
from dataclasses import dataclass, field
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from unittest.mock import patch

from checker.client import (
    ACTION_BUDGET,
    DEFAULT_PORT,
    MAX_RESPONSE_BYTES,
    REQUEST_TIMEOUT,
    Client,
)
from checker.protocol import Failure, Status

SESSION_TOKEN = "A" * 43
USER_BODY = json.dumps(
    {
        "user_id": "11111111-2222-3333-4444-555555555555",
        "username": "checker-user",
    }
).encode()


@dataclass
class Reply:
    status: int = 200
    body: bytes = b'{"ok":true}'
    headers: list[tuple[str, str]] = field(default_factory=list)
    content_type: str | None = "application/json"
    delay: float = 0


@dataclass
class ServerState:
    replies: deque[Reply]
    requests: list[dict] = field(default_factory=list)


def make_handler(state: ServerState):
    class Handler(BaseHTTPRequestHandler):
        protocol_version = "HTTP/1.1"

        def do_GET(self):
            self.respond()

        def do_POST(self):
            self.respond()

        def do_PUT(self):
            self.respond()

        def respond(self):
            length = int(self.headers.get("Content-Length", "0"))
            request = {
                "method": self.command,
                "path": self.path,
                "headers": dict(self.headers.items()),
                "body": self.rfile.read(length),
            }
            state.requests.append(request)
            reply = state.replies.popleft() if state.replies else Reply(status=500)
            if reply.delay:
                time.sleep(reply.delay)
            self.send_response(reply.status)
            if reply.content_type is not None:
                self.send_header("Content-Type", reply.content_type)
            for name, value in reply.headers:
                self.send_header(name, value)
            self.send_header("Content-Length", str(len(reply.body)))
            self.send_header("Connection", "close")
            self.end_headers()
            try:
                self.wfile.write(reply.body)
            except OSError:
                pass
            self.close_connection = True

        def log_message(self, format, *args):
            pass

    return Handler


class ClientTests(unittest.TestCase):
    def start_server(self, *replies: Reply):
        state = ServerState(deque(replies))
        server = ThreadingHTTPServer(("127.0.0.1", 0), make_handler(state))
        server.daemon_threads = True
        thread = threading.Thread(
            target=server.serve_forever, kwargs={"poll_interval": 0.01}, daemon=True
        )
        thread.start()
        self.addCleanup(server.server_close)
        self.addCleanup(thread.join, 1)
        self.addCleanup(server.shutdown)
        return state, f"127.0.0.1:{server.server_port}"

    def assert_failure(self, callback, status, secrets=()):
        with self.assertRaises(Failure) as raised:
            callback()
        failure = raised.exception
        self.assertEqual(failure.status, status)
        rendered = " ".join((failure.public, failure.private, str(failure)))
        for secret in secrets:
            self.assertNotIn(secret, rendered)

    def test_network_limits_match_checker_budget(self):
        self.assertEqual(DEFAULT_PORT, 8080)
        self.assertEqual(REQUEST_TIMEOUT, 2)
        self.assertEqual(ACTION_BUDGET, 15)
        self.assertEqual(MAX_RESPONSE_BYTES, 64 * 1024)

    def test_parses_host_port_and_ipv6_forms(self):
        self.assertEqual(Client.address("localhost"), ("localhost", 8080))
        self.assertEqual(Client.address("service.local:9090"), ("service.local", 9090))
        self.assertEqual(Client.address("127.0.0.1:8081"), ("127.0.0.1", 8081))
        self.assertEqual(Client.address("::1"), ("::1", 8080))
        self.assertEqual(Client.address("[::1]:9091"), ("::1", 9091))

    def test_rejects_invalid_host_without_echoing_credentials(self):
        self.assert_failure(
            lambda: Client.address("user:secret@service.local"),
            Status.CHECK_FAILED,
            secrets=("secret",),
        )
        for host in (
            "",
            "service.local:0",
            "service.local:bad",
            "[::1",
            "http://service.local",
        ):
            with self.subTest(host=host):
                self.assert_failure(
                    lambda host=host: Client.address(host), Status.CHECK_FAILED
                )

    def test_request_json_sends_payload_auth_and_custom_headers(self):
        state, host = self.start_server(Reply(status=201, body=b'{"created":true}'))
        result = Client(host).request_json(
            "POST",
            "/api/v1/cases",
            201,
            payload={"origin": "Moscow"},
            cookie="session=session-token",
            bearer="license-token",
            headers={"X-Forwarded-Role": "approver"},
        )

        self.assertEqual(result, {"created": True})
        request = state.requests[0]
        self.assertEqual(request["method"], "POST")
        self.assertEqual(request["path"], "/api/v1/cases")
        self.assertEqual(json.loads(request["body"]), {"origin": "Moscow"})
        self.assertEqual(request["headers"]["Content-Type"], "application/json")
        self.assertEqual(request["headers"]["Cookie"], "session=session-token")
        self.assertEqual(request["headers"]["Authorization"], "Bearer license-token")
        self.assertEqual(request["headers"]["X-Forwarded-Role"], "approver")

    def test_register_validates_user_response(self):
        _, host = self.start_server(Reply(status=201, body=USER_BODY))

        result = Client(host).register("checker-user", "checker-password")

        self.assertEqual(result["username"], "checker-user")

    def test_register_rejects_invalid_user_schema(self):
        _, host = self.start_server(Reply(status=201, body=b'{"username":4}'))

        self.assert_failure(
            lambda: Client(host).register("checker-user", "checker-password"),
            Status.MUMBLE,
            secrets=("checker-password",),
        )

    def test_login_returns_only_the_session_cookie_pair(self):
        state, host = self.start_server(
            Reply(
                body=USER_BODY,
                headers=[
                    ("Set-Cookie", "preference=dark; Path=/"),
                    (
                        "Set-Cookie",
                        f"session={SESSION_TOKEN}; Path=/; HttpOnly; SameSite=Lax",
                    ),
                ],
            )
        )

        user, cookie = Client(host).login("checker-user", "checker-password")

        self.assertEqual(user["username"], "checker-user")
        self.assertEqual(cookie, f"session={SESSION_TOKEN}")
        self.assertEqual(state.requests[0]["path"], "/api/v1/login")
        self.assertNotIn("checker-password", cookie)

    def test_login_rejects_missing_duplicate_and_invalid_session_cookie(self):
        replies = (
            Reply(body=USER_BODY, headers=[("Set-Cookie", "other=value; Path=/")]),
            Reply(
                body=USER_BODY,
                headers=[
                    ("Set-Cookie", f"session={SESSION_TOKEN}; Path=/"),
                    ("Set-Cookie", f"session={SESSION_TOKEN}; Path=/"),
                ],
            ),
            Reply(
                body=USER_BODY, headers=[("Set-Cookie", "session=unsafe/value; Path=/")]
            ),
        )
        state, host = self.start_server(*replies)
        client = Client(host)

        for _ in replies:
            self.assert_failure(
                lambda: client.login("checker-user", "checker-password"),
                Status.MUMBLE,
                secrets=("checker-password", SESSION_TOKEN),
            )
        self.assertEqual(len(state.requests), 3)

    def test_response_statuses_map_to_checker_verdicts(self):
        cases = (
            (429, False, Status.DOWN),
            (503, False, Status.DOWN),
            (401, True, Status.CORRUPT),
            (403, True, Status.CORRUPT),
            (404, True, Status.CORRUPT),
            (410, True, Status.CORRUPT),
            (404, False, Status.MUMBLE),
            (409, True, Status.MUMBLE),
        )
        state, host = self.start_server(
            *(
                Reply(status=status, body=b'{"message":"flag=secret"}')
                for status, _, _ in cases
            )
        )
        client = Client(host)

        for status, stored, expected_status in cases:
            with self.subTest(status=status, stored=stored):
                self.assert_failure(
                    lambda stored=stored: client.request_json(
                        "GET", "/api/v1/item", 200, stored=stored
                    ),
                    expected_status,
                    secrets=("flag=secret",),
                )
        self.assertEqual(len(state.requests), len(cases))

    def test_invalid_content_type_json_and_object_schema_are_mumble(self):
        replies = (
            Reply(content_type="text/html", body=b"<html>secret</html>"),
            Reply(body=b"not-json"),
            Reply(body=b"[]"),
            Reply(body=b'{"item":1,"item":2}'),
            Reply(body=b'{"value":NaN}'),
        )
        _, host = self.start_server(*replies)
        client = Client(host)

        for _ in replies:
            self.assert_failure(
                lambda: client.request_json("GET", "/api/v1/item", 200),
                Status.MUMBLE,
                secrets=("secret",),
            )

    def test_response_larger_than_limit_is_mumble(self):
        body = b'{"data":"' + b"x" * MAX_RESPONSE_BYTES + b'"}'
        _, host = self.start_server(Reply(body=body))

        self.assert_failure(
            lambda: Client(host).request_json("GET", "/api/v1/item", 200),
            Status.MUMBLE,
        )

    def test_redirect_is_not_followed_and_retry_is_not_attempted(self):
        state, host = self.start_server(
            Reply(status=302, headers=[("Location", "/next")]),
        )

        self.assert_failure(
            lambda: Client(host).request_json("GET", "/start", 200),
            Status.MUMBLE,
        )
        self.assertEqual(len(state.requests), 1)

    def test_malformed_http_response_is_mumble(self):
        class MalformedHandler(BaseHTTPRequestHandler):
            def do_GET(self):
                self.connection.sendall(b"not an HTTP response\r\n\r\n")

            def log_message(self, format, *args):
                pass

        server = ThreadingHTTPServer(("127.0.0.1", 0), MalformedHandler)
        server.daemon_threads = True
        thread = threading.Thread(
            target=server.serve_forever, kwargs={"poll_interval": 0.01}, daemon=True
        )
        thread.start()
        self.addCleanup(server.server_close)
        self.addCleanup(thread.join, 1)
        self.addCleanup(server.shutdown)

        self.assert_failure(
            lambda: Client(f"127.0.0.1:{server.server_port}").request_json(
                "GET", "/", 200
            ),
            Status.MUMBLE,
        )

    def test_transport_timeout_is_down(self):
        _, host = self.start_server(Reply(delay=0.3))
        with patch("checker.client.REQUEST_TIMEOUT", 0.08):
            self.assert_failure(
                lambda: Client(host).request_json("GET", "/slow", 200),
                Status.DOWN,
            )

    def test_action_deadline_applies_across_requests(self):
        state, host = self.start_server(Reply(delay=0.06), Reply(delay=0.2))
        with patch("checker.client.ACTION_BUDGET", 0.15):
            client = Client(host)
        client.request_json("GET", "/first", 200)

        self.assert_failure(
            lambda: client.request_json("GET", "/second", 200),
            Status.DOWN,
        )
        self.assertEqual(len(state.requests), 2)

    def test_action_deadline_prevents_a_new_request(self):
        state, host = self.start_server(Reply())
        with patch("checker.client.ACTION_BUDGET", 0.02):
            client = Client(host)
        time.sleep(0.04)

        self.assert_failure(
            lambda: client.request_json("GET", "/expired", 200),
            Status.DOWN,
        )
        self.assertEqual(state.requests, [])


if __name__ == "__main__":
    unittest.main()
