#!/usr/bin/env python3
"""ForcAD checker actions for Wahrwelt."""

from __future__ import annotations

import base64
import re
import secrets
import sys
import time

try:
    from .client import Client
    from .protocol import PLACE_FIELDS, Arguments, Failure, State, Status
except ImportError:
    from client import Client
    from protocol import PLACE_FIELDS, Arguments, Failure, State, Status


PUBLIC_HINT = True

_UUID = re.compile(
    r"[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-"
    r"[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}\Z"
)
_NUMBER_PREFIXES = {
    1: "MNF",
    2: "CRT",
    3: "INV",
    4: "INSP",
    5: "LIC",
    6: "CTR",
    8: "QUO",
    9: "AWB",
    10: "GATE",
}
_PLACE_PATHS = {
    1: "/api/v1/manifests",
    2: "/api/v1/certificates",
    3: "/api/v1/invoices",
    4: "/api/v1/inspections",
    5: "/api/v1/licenses",
    6: "/api/v1/contracts",
    8: "/api/v1/quotas",
    9: "/api/v1/archive-waybills",
    10: "/api/v1/gate-passes",
}
_EPS_DOCUMENT = b"%!PS-Adobe-3.0 EPSF-3.0\n%%BoundingBox: 0 0 1 1\nshowpage\n"
_PREVIEW_DELAYS = (0.1, 0.2, 0.4, 0.8, 1.0)


def run_put(client: Client, private: str, flag: str, place: int) -> State:
    """Create a flag object and prove that the owner can read the exact value."""
    if place not in PLACE_FIELDS:
        raise Failure(
            Status.CHECK_FAILED, "Checker invocation failed", "run_put - invalid place"
        )

    username, password, cookie = _create_owner(client)
    revision_id = None

    if place == 1:
        result = client.request_json(
            "POST",
            _PLACE_PATHS[1],
            201,
            {
                "origin": "North Bastion",
                "destination": "South Bastion",
                "sealed_route": flag,
            },
            cookie=cookie,
        )
        object_number = _number(result, "manifest_number", "MNF")
        revision_id = _uuid(result, "revision_id")
    elif place == 2:
        result = client.request_json(
            "POST", _PLACE_PATHS[2], 201, {"origin_phrase": flag}, cookie=cookie
        )
        object_number = _number(result, "serial", "CRT")
    elif place == 3:
        result = client.request_json(
            "POST",
            _PLACE_PATHS[3],
            201,
            {"amount": 700, "payment_ref": flag},
            cookie=cookie,
        )
        object_number = _number(result, "invoice_number", "INV")
    elif place == 4:
        case_id = _create_case(client, cookie, "service")
        result = client.request_json(
            "POST",
            _PLACE_PATHS[4],
            201,
            {
                "case_id": case_id,
                "inspector_username": "inspector",
                "inspector_note": flag,
            },
            cookie=cookie,
        )
        object_number = _number(result, "inspection_number", "INSP")
    elif place == 5:
        result = client.request_json(
            "POST", _PLACE_PATHS[5], 201, {"release_pin": flag}, cookie=cookie
        )
        object_number = _number(result, "license_number", "LIC")
    elif place == 6:
        result = client.request_json(
            "POST",
            _PLACE_PATHS[6],
            201,
            _contract_payload(flag),
            cookie=cookie,
        )
        object_number = _number(result, "contract_number", "CTR")
    elif place == 7:
        case_id = _create_case(client, cookie, "service")
        result = client.request_json(
            "POST",
            f"/api/v1/cases/{case_id}/route-slots",
            201,
            {
                "carrier_username": "carrier",
                "inspector_username": "inspector",
                "slot_code": flag,
            },
            cookie=cookie,
        )
        object_number = _uuid(result, "route_slot_id")
    elif place == 8:
        result = client.request_json(
            "POST", _PLACE_PATHS[8], 201, _quota_payload(flag), cookie=cookie
        )
        object_number = _number(result, "quota_number", "QUO")
    elif place == 9:
        result = client.request_json(
            "POST", _PLACE_PATHS[9], 201, {"archive_code": flag}, cookie=cookie
        )
        object_number = _number(result, "waybill_number", "AWB")
    else:
        result = client.request_json(
            "POST",
            _PLACE_PATHS[10],
            201,
            {
                "gate_operator_username": "inspector",
                "state": "issued",
                "gate_word": flag,
            },
            cookie=cookie,
        )
        object_number = _number(result, "gate_pass_number", "GATE")

    state = State(place, username, password, object_number, revision_id)
    _read_flag(client, state, flag, stored=False)
    return state


def run_get(client: Client, raw_state: str, flag: str, place: int) -> str:
    state = State.decode(raw_state, place)
    _validate_state_identifier(state)
    _read_flag(client, state, flag, stored=True)
    return state.object_number


def run_check(client: Client) -> None:
    health = client.request_json("GET", "/healthz", 200)
    _equal(health, "status", "ok", "health response")

    owner, owner_password = _register_owner(client)
    owner_user, owner_cookie = _login(client, owner, owner_password)
    carrier_user, carrier_cookie = _login(client, "carrier", "carrier")
    inspector_user, inspector_cookie = _login(client, "inspector", "inspector")
    if (
        len({owner_user["user_id"], carrier_user["user_id"], inspector_user["user_id"]})
        != 3
    ):
        _invalid("checker accounts")

    case_id = _create_case(client, owner_cookie, "service")
    case = client.request_json(
        "GET", f"/api/v1/cases/{case_id}", 200, cookie=owner_cookie
    )
    _equal(case, "case_id", case_id, "case response")

    manifest = client.request_json(
        "POST",
        "/api/v1/manifests",
        201,
        {
            "origin": "North Bastion",
            "destination": "South Bastion",
            "sealed_route": "check-route",
        },
        cookie=owner_cookie,
    )
    manifest_number = _number(manifest, "manifest_number", "MNF")
    revision_id = _uuid(manifest, "revision_id")
    manifest_read = client.request_json(
        "GET",
        f"/api/v1/manifests/{manifest_number}/revisions/{revision_id}",
        200,
        cookie=owner_cookie,
        stored=True,
    )
    _equal(manifest_read, "manifest_number", manifest_number, "manifest response")
    _equal(manifest_read, "revision_id", revision_id, "manifest response")
    _equal(manifest_read, "sealed_route", "check-route", "manifest response")

    license_result = client.request_json(
        "POST",
        "/api/v1/licenses",
        201,
        {"release_pin": "check-pin"},
        cookie=owner_cookie,
    )
    license_number = _number(license_result, "license_number", "LIC")
    token_result = client.request_json(
        "POST",
        f"/api/v1/licenses/{license_number}/tokens",
        201,
        {"scope": "license:read"},
        cookie=owner_cookie,
    )
    token = _text(token_result, "token")
    _text(token_result, "expires_at")
    pin_result = client.request_json(
        "GET",
        f"/api/v1/licenses/{license_number}/release-pin",
        200,
        bearer=token,
        stored=True,
    )
    _equal(pin_result, "license_number", license_number, "license response")
    _equal(pin_result, "release_pin", "check-pin", "license response")

    contract_result = client.request_json(
        "POST",
        "/api/v1/contracts",
        201,
        _contract_payload("check-crew"),
        cookie=owner_cookie,
    )
    contract_number = _number(contract_result, "contract_number", "CTR")
    contract = client.request_json(
        "GET",
        f"/api/v1/contracts/{contract_number}",
        200,
        cookie=owner_cookie,
        stored=True,
    )
    _equal(contract, "contract_number", contract_number, "contract response")
    _equal(contract, "crew_phrase", "check-crew", "contract response")

    inspection_result = client.request_json(
        "POST",
        "/api/v1/inspections",
        201,
        {
            "case_id": case_id,
            "inspector_username": "inspector",
            "inspector_note": "check-note",
        },
        cookie=owner_cookie,
    )
    inspection_number = _number(inspection_result, "inspection_number", "INSP")
    note = client.request_json(
        "GET",
        f"/api/v1/inspections/{inspection_number}/note",
        200,
        cookie=inspector_cookie,
        stored=True,
    )
    _equal(note, "inspector_note", "check-note", "inspection response")

    gate_result = client.request_json(
        "POST",
        "/api/v1/gate-passes",
        201,
        {
            "gate_operator_username": "inspector",
            "state": "issued",
            "gate_word": "check-gate",
        },
        cookie=owner_cookie,
    )
    gate_number = _number(gate_result, "gate_pass_number", "GATE")
    gate = client.request_json(
        "GET",
        f"/api/v1/gate-passes/{gate_number}",
        200,
        cookie=inspector_cookie,
        stored=True,
    )
    _equal(gate, "gate_pass_number", gate_number, "gate pass response")
    _equal(gate, "gate_word", "check-gate", "gate pass response")

    slot_result = client.request_json(
        "POST",
        f"/api/v1/cases/{case_id}/route-slots",
        201,
        {
            "carrier_username": "carrier",
            "inspector_username": "inspector",
            "slot_code": "check-slot",
        },
        cookie=owner_cookie,
    )
    slot_id = _uuid(slot_result, "route_slot_id")
    first_approval = client.request_json(
        "POST",
        f"/api/v1/route-slots/{slot_id}/approvals",
        200,
        {"role": "carrier"},
        cookie=carrier_cookie,
    )
    _approval_state(first_approval, False)
    second_approval = client.request_json(
        "POST",
        f"/api/v1/route-slots/{slot_id}/approvals",
        200,
        {"role": "inspector"},
        cookie=inspector_cookie,
    )
    _approval_state(second_approval, True)
    slot = client.request_json(
        "GET",
        f"/api/v1/route-slots/{slot_id}/code",
        200,
        cookie=inspector_cookie,
        stored=True,
    )
    _equal(slot, "slot_code", "check-slot", "route slot response")

    _check_preview(client, owner_cookie)


def _read_flag(client: Client, state: State, flag: str, stored: bool) -> None:
    _, cookie = _login(client, state.username, state.password)
    place = state.place
    if place == 1:
        path = f"/api/v1/manifests/{state.object_number}/revisions/{state.revision_id}"
        data = client.request_json("GET", path, 200, cookie=cookie, stored=stored)
        _stored_equal(data, "manifest_number", state.object_number, stored)
        _stored_equal(data, "revision_id", state.revision_id, stored)
    elif place == 2:
        data = client.request_json(
            "GET",
            f"/api/v1/certificates/{state.object_number}",
            200,
            cookie=cookie,
            stored=stored,
        )
    elif place == 3:
        data = client.request_json(
            "GET",
            f"/api/v1/invoices/{state.object_number}",
            200,
            cookie=cookie,
            stored=stored,
        )
    elif place == 4:
        data = client.request_json(
            "GET",
            f"/api/v1/inspections/{state.object_number}/note",
            200,
            cookie=cookie,
            stored=stored,
        )
    elif place == 5:
        token_result = client.request_json(
            "POST",
            f"/api/v1/licenses/{state.object_number}/tokens",
            201,
            {"scope": "license:read"},
            cookie=cookie,
            stored=stored,
        )
        token = _text(token_result, "token", stored=stored)
        _text(token_result, "expires_at", stored=stored)
        data = client.request_json(
            "GET",
            f"/api/v1/licenses/{state.object_number}/release-pin",
            200,
            bearer=token,
            stored=stored,
        )
        _stored_equal(data, "license_number", state.object_number, stored)
    elif place == 6:
        data = client.request_json(
            "GET",
            f"/api/v1/contracts/{state.object_number}",
            200,
            cookie=cookie,
            stored=stored,
        )
    elif place == 7:
        data = client.request_json(
            "GET",
            f"/api/v1/route-slots/{state.object_number}/code",
            200,
            cookie=cookie,
            stored=stored,
        )
    elif place == 8:
        data = client.request_json(
            "GET",
            f"/api/v1/quotas/{state.object_number}",
            200,
            cookie=cookie,
            stored=stored,
        )
    elif place == 9:
        data = client.request_json(
            "GET",
            f"/api/v1/archive-waybills/{state.object_number}",
            200,
            cookie=cookie,
            stored=stored,
        )
    elif place == 10:
        data = client.request_json(
            "GET",
            f"/api/v1/gate-passes/{state.object_number}",
            200,
            cookie=cookie,
            stored=stored,
        )
    else:
        raise Failure(
            Status.CHECK_FAILED,
            "Checker invocation failed",
            "_read_flag - invalid place",
        )

    field = PLACE_FIELDS[place]
    if place == 2:
        _stored_equal(data, "serial", state.object_number, stored)
    elif place == 3:
        _stored_equal(data, "invoice_number", state.object_number, stored)
    elif place == 4 or place == 5:
        pass
    elif place == 6:
        _stored_equal(data, "contract_number", state.object_number, stored)
    elif place == 7:
        _stored_equal(data, "slot_code", flag, stored)
        return
    elif place == 8:
        _stored_equal(data, "quota_number", state.object_number, stored)
    elif place == 9:
        _stored_equal(data, "waybill_number", state.object_number, stored)
    elif place == 10:
        _stored_equal(data, "gate_pass_number", state.object_number, stored)
    _stored_equal(data, field, flag, stored)


def _create_owner(client: Client) -> tuple[str, str, str]:
    username, password = _register_owner(client)
    _, cookie = _login(client, username, password)
    return username, password, cookie


def _register_owner(client: Client) -> tuple[str, str]:
    username = "ww" + secrets.token_hex(6)
    password = secrets.token_urlsafe(24)
    user = client.register(username, password)
    _user(user, username)
    return username, password


def _login(client: Client, username: str, password: str) -> tuple[dict, str]:
    user, cookie = client.login(username, password)
    _user(user, username)
    if not isinstance(cookie, str) or not cookie.startswith("session="):
        _invalid("session response")
    return user, cookie


def _user(data: dict, username: str) -> None:
    _uuid(data, "user_id")
    _equal(data, "username", username, "user response")


def _validate_state_identifier(state: State) -> None:
    if state.place == 1 and state.revision_id is None:
        raise Failure(
            Status.CHECK_FAILED,
            "Checker state is invalid",
            "checker - manifest state lacks a revision identifier",
        )
    if state.place == 7:
        if not _UUID.fullmatch(state.object_number):
            raise Failure(
                Status.CHECK_FAILED,
                "Checker state is invalid",
                "checker - invalid stored object identifier",
            )
        return
    prefix = _NUMBER_PREFIXES[state.place]
    if not re.fullmatch(
        rf"{re.escape(prefix)}-{_UUID.pattern[:-2]}", state.object_number
    ):
        raise Failure(
            Status.CHECK_FAILED,
            "Checker state is invalid",
            "checker - invalid stored object identifier",
        )


def _create_case(client: Client, cookie: str, kind: str) -> str:
    data = client.request_json(
        "POST",
        "/api/v1/cases",
        201,
        {"kind": kind, "origin": "North Bastion", "destination": "South Bastion"},
        cookie=cookie,
    )
    _equal(data, "kind", kind, "case response")
    return _uuid(data, "case_id")


def _contract_payload(flag: str) -> dict:
    return {
        "crew_name": "North Repair Crew",
        "service": "bridge repair",
        "crew_phrase": flag,
        "starts_at": "2030-01-01T09:00:00Z",
        "ends_at": "2030-01-01T17:00:00Z",
    }


def _quota_payload(flag: str) -> dict:
    expires_at = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime(time.time() + 86400))
    return {"quota_code": flag, "volume": 700, "expires_at": expires_at}


def _check_preview(client: Client, cookie: str) -> None:
    created = client.request_json("POST", "/api/v1/previews", 201, cookie=cookie)
    preview_id = _uuid(created, "preview_id")
    _equal(created, "state", "created", "preview response")
    encoded = base64.b64encode(_EPS_DOCUMENT).decode("ascii")
    uploaded = client.request_json(
        "PUT",
        f"/api/v1/previews/{preview_id}",
        202,
        {"eps_base64": encoded},
        cookie=cookie,
    )
    _equal(uploaded, "preview_id", preview_id, "preview response")
    _equal(uploaded, "state", "queued", "preview response")

    for delay in _PREVIEW_DELAYS:
        result = client.request_json(
            "GET", f"/api/v1/previews/{preview_id}", 200, cookie=cookie, stored=True
        )
        _equal(result, "preview_id", preview_id, "preview response")
        state = _text(result, "state")
        if state == "complete":
            encoded_png = _text(result, "png_base64")
            try:
                png = base64.b64decode(encoded_png, validate=True)
            except (ValueError, TypeError):
                _invalid("preview image")
            if len(png) <= 8 or not png.startswith(b"\x89PNG\r\n\x1a\n"):
                _invalid("preview image")
            return
        if state == "failed":
            _invalid("preview result")
        if state not in {"queued", "processing"}:
            _invalid("preview state")
        time.sleep(delay)
    _invalid("preview did not complete")


def _approval_state(data: dict, ready: bool) -> None:
    _boolean(data, "carrier_approved", "approval response")
    _boolean(data, "inspector_approved", "approval response")
    _equal(data, "ready", ready, "approval response")
    if data["carrier_approved"] is not True:
        _invalid("approval response")
    if data["inspector_approved"] is not ready:
        _invalid("approval response")


def _number(data: dict, key: str, prefix: str) -> str:
    value = _text(data, key)
    if not re.fullmatch(rf"{re.escape(prefix)}-{_UUID.pattern[:-2]}", value):
        _invalid(f"{prefix} response")
    return value


def _uuid(data: dict, key: str) -> str:
    value = _text(data, key)
    if not _UUID.fullmatch(value):
        _invalid("identifier response")
    return value


def _text(data: dict, key: str, stored: bool = False) -> str:
    value = data.get(key)
    if type(value) is not str or not value:
        _invalid("stored object" if stored else "service response", stored)
    return value


def _boolean(data: dict, key: str, context: str) -> bool:
    value = data.get(key)
    if type(value) is not bool:
        _invalid(context)
    return value


def _equal(data: dict, key: str, expected, context: str) -> None:
    if (
        key not in data
        or type(data[key]) is not type(expected)
        or data[key] != expected
    ):
        _invalid(context)


def _stored_equal(data: dict, key: str, expected, stored: bool) -> None:
    if (
        key not in data
        or type(data[key]) is not type(expected)
        or data[key] != expected
    ):
        _invalid("stored object" if stored else "service response", stored)


def _invalid(context: str, stored: bool = False) -> None:
    status = Status.CORRUPT if stored else Status.MUMBLE
    message = "Stored flag is unavailable" if stored else "Invalid service response"
    raise Failure(status, message, f"checker - invalid {context}")


def main(args: list[str] | None = None) -> int:
    try:
        arguments = Arguments.parse(sys.argv[1:] if args is None else args)
        client = Client(arguments.host)
        if arguments.action == "put":
            state = run_put(client, arguments.private, arguments.flag, arguments.place)
            print(state.object_number)
            print(state.encode(), file=sys.stderr)
        elif arguments.action == "get":
            run_get(client, arguments.private, arguments.flag, arguments.place)
            print("OK")
        else:
            run_check(client)
            print("OK")
        return int(Status.UP)
    except Failure as failure:
        print(failure.public)
        print(failure.private, file=sys.stderr)
        return int(failure.status)
    except Exception:  # noqa: BLE001
        print("Checker failed")
        print("checker - unexpected internal error", file=sys.stderr)
        return int(Status.CHECK_FAILED)


if __name__ == "__main__":
    raise SystemExit(main())
