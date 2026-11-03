#!/usr/bin/env python3
"""Repeatable black-box scenario; each run creates its own problem. Standard library only."""

import json
import os
import time
import urllib.request
import urllib.error

BASE = os.environ.get("BASE_URL", "http://localhost:8080").rstrip("/")
ROLES = {
    "region-tomsk": "region",
    "region-other": "region",
    "tech-water": "technologist",
    "expert-1": "expert",
    "investor-a": "investor",
    "investor-b": "investor",
    "investor-c": "investor",
}


def call(method, path, user="region-tomsk", body=None, expected=200, role=None):
    """Send an authenticated request and assert its expected HTTP status."""
    req = urllib.request.Request(
        BASE + "/api" + path,
        data=json.dumps(body).encode() if body is not None else None,
        headers={
            "Content-Type": "application/json",
            "X-User-Id": user,
            "X-Role": role or ROLES[user],
        },
        method=method,
    )
    try:
        with urllib.request.urlopen(req, timeout=10) as response:
            code, data = response.status, json.load(response)
    except urllib.error.HTTPError as exc:
        code, data = exc.code, json.load(exc)
    assert code == expected, (path, code, data, expected)
    return data


pid = call("POST", "/problems", body={"text": "Томская область, очистка воды"})[
    "problem_id"
]
p = "/problems/" + pid
call("POST", "/problems", "tech-water", {"text": "Запрещено"}, 403)
call("POST", p + "/pool", body={"amount": 1000000}, expected=409)
call(
    "POST", p + "/investments", "investor-a", {"amount": 1, "key": "before-match"}, 409
)
call("POST", p + "/solutions", "investor-a", {"text": "Запрещено"}, 403)
call("POST", p + "/match", body={"solution_id": "none"}, expected=403)
call("POST", "/problems", "expert-1", {"text": "Подмена роли"}, 403, role="region")
solution = call("POST", p + "/solutions", "tech-water", {"text": "Модульная фильтрация воды"})
match = {"solution_id": solution["solution_id"]}
assert call("POST", p + "/match", "expert-1", match)["replayed"] is False
assert call("POST", p + "/match", "expert-1", match)["replayed"] is True
call("POST", p + "/pool", body={"amount": 1000000})
profit = {"amount": 100000}
call("POST", p + "/profit", body=profit, expected=409)
for user, amount in [("investor-a", 600000), ("investor-b", 400000)]:
    investment = {"amount": amount, "key": "investment-" + user}
    call("POST", p + "/investments", user, investment)
    call("POST", p + "/investments", user, investment)
call("POST", p + "/investments", "investor-c", {"amount": 1, "key": "excess"}, 409)
payouts = call("POST", p + "/profit", body=profit)["distribution"]["payouts"]
assert payouts == [
    {"investor_id": "investor-a", "amount": 60000},
    {"investor_id": "investor-b", "amount": 40000},
], payouts
assert call("POST", p + "/profit", body=profit)["replayed"] is True  # no-op
call("POST", p + "/profit", body={"amount": 100001}, expected=409)
for user, amount in [("investor-a", 60000), ("investor-b", 40000)]:
    project = next(x for x in call("GET", "/state", user)["projects"] if x["id"] == pid)
    assert project["pool"]["status"] == "distributed"
    assert project["pool"]["distribution"]["payouts"] == [
        {"investor_id": user, "amount": amount}
    ]
    assert all(x["investor_id"] == user for x in project["pool"]["investments"])
    print(f"{user}: {amount}")
assert pid not in {x["id"] for x in call("GET", "/state", "region-other")["projects"]}
for _ in range(30):
    events = [e for e in call("GET", "/events") if e["problem_id"] == pid]
    if any(e["name"] == "ProfitDistributed" for e in events):
        break
    time.sleep(0.1)
counts = {
    name: sum(e["name"] == name for e in events)
    for name in [
        "ProblemPublished",
        "SolutionMatched",
        "InvestmentAccepted",
        "ProfitDistributed",
    ]
}
assert counts == {
    "ProblemPublished": 1,
    "SolutionMatched": 1,
    "InvestmentAccepted": 2,
    "ProfitDistributed": 1,
}, counts
print("PASS: happy-path, idempotency, roles, privacy, audit; problem=" + pid)
