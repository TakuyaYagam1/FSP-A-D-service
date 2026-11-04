"""Bounded HTTP client for the Wahrwelt checker."""

from __future__ import annotations

import http.client
import ipaddress
import json
import re
import socket
import time
from email import policy
from email.message import Message
from email.parser import Parser
from http.cookies import CookieError, SimpleCookie
from urllib.parse import urlsplit

try:
    from .protocol import Failure, Status
except ImportError:
    from protocol import Failure, Status


DEFAULT_PORT = 8080
REQUEST_TIMEOUT = 2.0
ACTION_BUDGET = 15.0
MAX_RESPONSE_BYTES = 64 * 1024
READ_CHUNK_BYTES = 4096

_HOSTNAME = re.compile(
    r"(?=.{1,253}\Z)(?:[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)"
    r"(?:\.(?:[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?))*\.?\Z"
)
_HTTP_TOKEN = re.compile(r"[!#$%&'*+.^_`|~0-9A-Za-z-]+\Z")
_SESSION_TOKEN = re.compile(r"[A-Za-z0-9_-]{43}\Z")
_UUID = re.compile(
    r"[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-"
    r"[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}\Z"
)


class Client:
    def __init__(self, host: str):
        self.host, self.port = self.address(host)
        self.deadline = time.monotonic() + ACTION_BUDGET

    @staticmethod
    def address(host: str) -> tuple[str, int]:
        try:
            if (
                not isinstance(host, str)
                or not host
                or host != host.strip()
                or any(ord(char) <= 32 or ord(char) == 127 for char in host)
            ):
                raise ValueError("invalid host")

            try:
                address = ipaddress.ip_address(host)
            except ValueError:
                address = None
            if address is not None:
                return str(address), DEFAULT_PORT

            if host.startswith("["):
                closing = host.find("]")
                if closing < 0:
                    raise ValueError("invalid IPv6 host")
                suffix = host[closing + 1 :]
                if suffix and not suffix.startswith(":"):
                    raise ValueError("invalid host suffix")
                if suffix == ":" or (suffix and not suffix[1:].isdecimal()):
                    raise ValueError("invalid port")
                ipaddress.IPv6Address(host[1:closing])
            elif host.count(":") == 1:
                _, separator, raw_port = host.rpartition(":")
                if not separator or not raw_port.isdecimal():
                    raise ValueError("invalid port")
            elif ":" in host:
                raise ValueError("IPv6 with a port must be bracketed")

            parts = urlsplit(f"//{host}")
            if (
                not parts.hostname
                or parts.username is not None
                or parts.password is not None
                or parts.path
                or parts.query
                or parts.fragment
            ):
                raise ValueError("invalid host")

            name = parts.hostname
            try:
                ipaddress.ip_address(name)
            except ValueError:
                if not _HOSTNAME.fullmatch(name):
                    raise ValueError("invalid hostname") from None

            port = parts.port if parts.port is not None else DEFAULT_PORT
            if port == 0:
                raise ValueError("invalid port")
            return name, port
        except (TypeError, ValueError):
            raise Failure(
                Status.CHECK_FAILED,
                "Checker invocation failed",
                "Client.address - expected a host or host:port",
            ) from None

    def timeout(self) -> float:
        remaining = self.deadline - time.monotonic()
        if remaining <= 0:
            raise TimeoutError("action deadline exceeded")
        return min(REQUEST_TIMEOUT, remaining)

    def request_json(
        self,
        method: str,
        path: str,
        expected: int,
        payload: dict | None = None,
        cookie: str | None = None,
        bearer: str | None = None,
        headers: dict[str, str] | None = None,
        stored: bool = False,
    ) -> dict:
        data, _ = self._request_json(
            method, path, expected, payload, cookie, bearer, headers, stored
        )
        return data

    def _request_json(
        self,
        method: str,
        path: str,
        expected: int,
        payload: dict | None = None,
        cookie: str | None = None,
        bearer: str | None = None,
        headers: dict[str, str] | None = None,
        stored: bool = False,
    ) -> tuple[dict, Message]:
        body, request_headers = self._prepare_request(
            method, path, expected, payload, cookie, bearer, headers
        )
        connection = http.client.HTTPConnection(self.host, self.port)
        try:
            connection.timeout = self.timeout()
            connection.connect()
            transport = connection.sock
            if transport is None:
                raise http.client.HTTPException("missing connection socket")
            transport.settimeout(self.timeout())
            connection.request(method, path, body=body, headers=request_headers)
            transport.settimeout(self.timeout())
            with connection.getresponse() as response:
                self.validate_status(response.status, expected, stored)
                self.validate_content_type(response)
                data = self.read_json(response, transport)
                return data, response.headers
        except TimeoutError as exc:
            raise Failure(
                Status.DOWN,
                "Service unavailable",
                f"Client.request_json - transport timeout ({type(exc).__name__})",
            ) from None
        except OSError as exc:
            raise Failure(
                Status.DOWN,
                "Service unavailable",
                f"Client.request_json - transport failure ({type(exc).__name__})",
            ) from None
        except http.client.HTTPException as exc:
            raise Failure(
                Status.MUMBLE,
                "Invalid service response",
                f"Client.request_json - invalid HTTP response ({type(exc).__name__})",
            ) from None
        finally:
            connection.close()

    @staticmethod
    def _prepare_request(
        method: str,
        path: str,
        expected: int,
        payload: dict | None,
        cookie: str | None,
        bearer: str | None,
        headers: dict[str, str] | None,
    ) -> tuple[bytes | None, dict[str, str]]:
        if (
            not isinstance(method, str)
            or not _HTTP_TOKEN.fullmatch(method)
            or not isinstance(path, str)
            or not path.startswith("/")
            or path.startswith("//")
            or any(ord(char) <= 32 or ord(char) == 127 for char in path)
            or "#" in path
            or type(expected) is not int
            or not 100 <= expected <= 599
        ):
            raise Failure(
                Status.CHECK_FAILED,
                "Checker invocation failed",
                "Client.request_json - invalid request parameters",
            )
        if payload is not None and not isinstance(payload, dict):
            raise Failure(
                Status.CHECK_FAILED,
                "Checker invocation failed",
                "Client.request_json - payload must be a JSON object",
            )

        try:
            body = (
                json.dumps(payload, separators=(",", ":"), ensure_ascii=False).encode(
                    "utf-8"
                )
                if payload is not None
                else None
            )
        except (TypeError, ValueError, UnicodeError, RecursionError):
            raise Failure(
                Status.CHECK_FAILED,
                "Checker invocation failed",
                "Client.request_json - payload is not valid JSON",
            ) from None

        request_headers = {"Accept": "application/json", "Connection": "close"}
        if headers is not None:
            if not isinstance(headers, dict):
                raise Failure(
                    Status.CHECK_FAILED,
                    "Checker invocation failed",
                    "Client.request_json - headers must be a mapping",
                )
            for name, value in headers.items():
                if (
                    not isinstance(name, str)
                    or not _HTTP_TOKEN.fullmatch(name)
                    or not isinstance(value, str)
                    or "\r" in value
                    or "\n" in value
                ):
                    raise Failure(
                        Status.CHECK_FAILED,
                        "Checker invocation failed",
                        "Client.request_json - invalid request header",
                    )
                request_headers[name] = value

        if body is not None:
            request_headers["Content-Type"] = "application/json"
        if cookie is not None:
            if not isinstance(cookie, str) or "\r" in cookie or "\n" in cookie:
                raise Failure(
                    Status.CHECK_FAILED,
                    "Checker invocation failed",
                    "Client.request_json - invalid cookie header",
                )
            request_headers["Cookie"] = cookie
        if bearer is not None:
            if not isinstance(bearer, str) or "\r" in bearer or "\n" in bearer:
                raise Failure(
                    Status.CHECK_FAILED,
                    "Checker invocation failed",
                    "Client.request_json - invalid bearer header",
                )
            request_headers["Authorization"] = f"Bearer {bearer}"
        return body, request_headers

    @staticmethod
    def validate_status(status: int, expected: int, stored: bool) -> None:
        if status == expected:
            return
        if stored and status in {401, 403, 404, 410}:
            raise Failure(
                Status.CORRUPT,
                "Stored flag is unavailable",
                f"Client.validate_status - stored object is inaccessible (HTTP {status})",
            )
        verdict = (
            Status.DOWN if status == 429 or 500 <= status <= 599 else Status.MUMBLE
        )
        raise Failure(
            verdict,
            "Unexpected service response",
            f"Client.validate_status - expected HTTP {expected}, received {status}",
        )

    @staticmethod
    def validate_content_type(response: http.client.HTTPResponse) -> None:
        values = response.headers.get_all("Content-Type", [])
        if len(values) != 1:
            raise Failure(
                Status.MUMBLE,
                "Invalid service response",
                "Client.validate_content_type - expected one JSON content type",
            )
        if "\r" in values[0] or "\n" in values[0]:
            raise Failure(
                Status.MUMBLE,
                "Invalid service response",
                "Client.validate_content_type - invalid content type header",
            )
        try:
            message = Parser(policy=policy.default).parsestr(
                f"Content-Type: {values[0]}\n\n"
            )
            content_type = message.get_content_type()
            charset = message.get_content_charset()
            defects = message.defects
        except (TypeError, ValueError):
            content_type = ""
            charset = ""
            defects = (True,)
        if (
            defects
            or content_type != "application/json"
            or (charset is not None and charset.lower() != "utf-8")
        ):
            raise Failure(
                Status.MUMBLE,
                "Invalid service response",
                "Client.validate_content_type - expected application/json UTF-8",
            )

    def read_json(
        self, response: http.client.HTTPResponse, transport: socket.socket
    ) -> dict:
        content_length = response.headers.get_all("Content-Length", [])
        if len(content_length) > 1:
            raise Failure(
                Status.MUMBLE,
                "Invalid service response",
                "Client.read_json - duplicate content length",
            )
        if content_length:
            raw_length = content_length[0]
            if not raw_length.isascii() or not raw_length.isdecimal():
                raise Failure(
                    Status.MUMBLE,
                    "Invalid service response",
                    "Client.read_json - invalid or oversized response length",
                )
            try:
                declared_length = int(raw_length)
            except ValueError:
                declared_length = MAX_RESPONSE_BYTES + 1
            if declared_length > MAX_RESPONSE_BYTES:
                raise Failure(
                    Status.MUMBLE,
                    "Invalid service response",
                    "Client.read_json - invalid or oversized response length",
                )
        transfer_encoding = response.headers.get_all("Transfer-Encoding", [])
        if content_length and transfer_encoding:
            raise Failure(
                Status.MUMBLE,
                "Invalid service response",
                "Client.read_json - conflicting response framing",
            )
        if transfer_encoding and (
            len(transfer_encoding) != 1 or transfer_encoding[0].lower() != "chunked"
        ):
            raise Failure(
                Status.MUMBLE,
                "Invalid service response",
                "Client.read_json - invalid transfer encoding",
            )

        content = bytearray()
        while len(content) <= MAX_RESPONSE_BYTES:
            if response.isclosed():
                break
            transport.settimeout(self.timeout())
            chunk = response.read1(
                min(READ_CHUNK_BYTES, MAX_RESPONSE_BYTES + 1 - len(content))
            )
            if not chunk:
                break
            content.extend(chunk)
        if len(content) > MAX_RESPONSE_BYTES:
            raise Failure(
                Status.MUMBLE,
                "Invalid service response",
                "Client.read_json - response exceeds size limit",
            )

        try:
            data = json.loads(
                content.decode("utf-8"),
                object_pairs_hook=self._unique_object,
                parse_constant=self._reject_constant,
            )
        except (UnicodeError, ValueError, TypeError, RecursionError):
            raise Failure(
                Status.MUMBLE,
                "Invalid service response",
                "Client.read_json - expected a UTF-8 JSON object",
            ) from None
        if not isinstance(data, dict):
            raise Failure(
                Status.MUMBLE,
                "Invalid service response",
                "Client.read_json - expected a JSON object",
            )
        return data

    @staticmethod
    def _unique_object(pairs: list[tuple[str, object]]) -> dict[str, object]:
        result: dict[str, object] = {}
        for key, value in pairs:
            if key in result:
                raise ValueError("duplicate JSON field")
            result[key] = value
        return result

    @staticmethod
    def _reject_constant(value: str) -> None:
        raise ValueError("invalid JSON constant")

    def register(self, username: str, password: str) -> dict:
        data, _ = self._request_json(
            "POST",
            "/api/v1/register",
            201,
            {"username": username, "password": password},
        )
        self.validate_user(data)
        return data

    def login(self, username: str, password: str) -> tuple[dict, str]:
        data, response_headers = self._request_json(
            "POST",
            "/api/v1/login",
            200,
            {"username": username, "password": password},
        )
        self.validate_user(data)
        return data, self.session_cookie(response_headers)

    @staticmethod
    def validate_user(data: dict) -> None:
        user_id = data.get("user_id")
        username = data.get("username")
        if (
            not isinstance(user_id, str)
            or not _UUID.fullmatch(user_id)
            or not isinstance(username, str)
            or not username
        ):
            raise Failure(
                Status.MUMBLE,
                "Invalid service response",
                "Client.validate_user - invalid user response schema",
            )

    @staticmethod
    def session_cookie(headers: Message) -> str:
        values = headers.get_all("Set-Cookie", [])
        sessions = []
        for header in values:
            first_pair = header.split(";", 1)[0].strip()
            name, separator, raw_value = first_pair.partition("=")
            if name != "session":
                continue
            if not separator or "," in raw_value:
                raise Failure(
                    Status.MUMBLE,
                    "Invalid service response",
                    "Client.session_cookie - invalid session cookie",
                )
            try:
                parsed = SimpleCookie()
                parsed.load(header)
                morsel = parsed.get("session")
            except (CookieError, TypeError, ValueError):
                morsel = None
            if morsel is None or not _SESSION_TOKEN.fullmatch(morsel.value):
                raise Failure(
                    Status.MUMBLE,
                    "Invalid service response",
                    "Client.session_cookie - invalid session cookie",
                )
            sessions.append(morsel.value)
        if len(sessions) != 1:
            raise Failure(
                Status.MUMBLE,
                "Invalid service response",
                "Client.session_cookie - expected one session cookie",
            )
        return f"session={sessions[0]}"
