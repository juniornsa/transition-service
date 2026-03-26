# CLAUDE.md — Observability Agent System

> Read this entire file before writing a single line of code.
> These are non-negotiable rules, not suggestions.

-----

## Project Identity

You are building a production-grade, multi-cluster Kubernetes observability agent system
using Google ADK. It is real infrastructure that will run against live clusters.

This codebase has zero tolerance for:

- Placeholder implementations
- Mock data in non-test code
- Deferred work marked TODO/FIXME
- Dead code of any kind
- Incomplete functions

-----

## The Non-Negotiable Rules

### Rule 1: No Deferred Implementations

NEVER write any of the following:

```python
# TODO: implement this later
# FIXME: add real logic here
# Coming soon
pass  # implement
raise NotImplementedError("TODO")
```

If you cannot implement something completely right now, STOP and tell the user
what is missing before writing any code. Do not scaffold an empty shell and
move on. Either implement it fully or don't write it at all.

-----

### Rule 2: No Mocks Outside of Tests

Mocks, stubs, and fake data belong ONLY in files under `tests/`.
They are NEVER acceptable in:

- `agents/`
- `tools/`
- `mcps/`
- `config/`
- Any production code path

**Banned patterns in production code:**

```python
# BANNED
return {"status": "ok", "data": "mock_response"}
return MagicMock()
data = {"fake": "data"}  # placeholder
result = "simulated_response"
time.sleep(1)  # simulating API call

# BANNED — if the MCP call isn't wired, don't fake it
def query_mimir(query: str) -> dict:
    return {"value": 0.5}  # placeholder until MCP is ready
```

**The only acceptable pattern when an integration isn't ready:**

```python
def query_mimir(query: str) -> dict:
    raise EnvironmentError(
        "Grafana MCP not configured. Set GRAFANA_MCP_ENDPOINT in environment. "
        "See docs/mcp-setup.md for deployment instructions."
    )
```

Fail loudly. Never fake silently.

-----

### Rule 3: No Dead Code

Every function, class, import, constant, and variable must be actively used.

After writing any code, before considering it done, you MUST verify:

- Every import is used
- Every function is called
- Every class is instantiated somewhere
- Every constant is referenced
- No commented-out code blocks exist

Run this mental audit explicitly. Do not skip it.

**Dead code patterns that are banned:**

```python
# Banned — unused import
import json  # never referenced below

# Banned — unused function
def _helper_that_was_replaced():
    pass

# Banned — commented-out old implementation
# def old_query():
#     return requests.get(...)

# Banned — unreachable branch
if False:
    do_something()

# Banned — unused variable
result = compute_something()  # result never used below
```

If you remove a function, check every call site. Remove the callers too if they
only existed to call the removed function. Trace the full removal.

-----

### Rule 4: Complete Function Contracts

Every function must:

1. Have a type signature (all parameters + return type)
1. Have a docstring describing what it does, not what it is
1. Handle its error cases explicitly — no bare `except: pass`
1. Return real data or raise a typed exception

```python
# WRONG
def get_cluster_baselines(cluster_id):
    # TODO: query Redis
    pass

# WRONG
def get_cluster_baselines(cluster_id):
    try:
        return redis.get(f"baselines:{cluster_id}")
    except:
        return {}

# RIGHT
async def get_cluster_baselines(cluster_id: str) -> ClusterBaselines:
    """
    Load per-service metric baselines for a cluster from Redis.
    Baselines are seeded from 30 days of Mimir data and refreshed weekly.

    Raises:
        RedisConnectionError: if Redis is unreachable
        KeyError: if cluster_id has no baseline entry (cluster not onboarded)
    """
    key = f"context:baselines:{cluster_id}"
    raw = await redis_client.get(key)
    if raw is None:
        raise KeyError(
            f"No baseline found for cluster '{cluster_id}'. "
            f"Run 'scripts/seed_baselines.py --cluster {cluster_id}' to initialize."
        )
    return ClusterBaselines.model_validate_json(raw)
```

-----

### Rule 5: No Partial Agent Implementations

An ADK agent is not complete until:

- All declared tools are fully implemented and callable
- Session state reads/writes are wired to real Redis keys
- MCP client connections are initialized, not assumed
- The agent can be instantiated and run without crashing on a real environment
- Error handling covers: MCP timeout, Redis miss, malformed API response

**Banned agent pattern:**

```python
class MetricsAgent(LlmAgent):
    def __init__(self):
        super().__init__(
            name="MetricsAgent",
            tools=[self.query_mimir, self.write_findings]
        )

    async def query_mimir(self, query: str) -> str:
        # TODO: connect to Grafana MCP
        return "placeholder metrics data"

    async def write_findings(self, findings: str) -> str:
        # TODO: write to Redis
        print(f"Would write: {findings}")
        return "ok"
```

**Required pattern:**

```python
class MetricsAgent(LlmAgent):
    def __init__(self, grafana_mcp: GrafanaMCPClient, redis: RedisClient):
        self._grafana = grafana_mcp
        self._redis = redis
        super().__init__(
            name="MetricsAgent",
            tools=[self.query_prometheus, self.query_histogram_percentiles, self.write_findings]
        )

    async def query_prometheus(
        self,
        query: str,
        cluster_id: str,
        start: datetime,
        end: datetime
    ) -> PrometheusQueryResult:
        """Execute PromQL range query against Mimir for the given cluster."""
        return await self._grafana.query_prometheus(
            query=query,
            datasource=await self._get_datasource(cluster_id),
            start=start,
            end=end
        )
```

-----

### Rule 6: Strict Dependency Injection — No Globals, No Singletons

Agents and tools receive their dependencies via constructor. Never:

```python
# BANNED
from config import redis_client  # global singleton

async def write_findings(findings: dict):
    await redis_client.set(...)  # where did this come from?
```

Always:

```python
# REQUIRED
class LogsAgent(LlmAgent):
    def __init__(self, grafana_mcp: GrafanaMCPClient, redis: RedisClient):
        self._grafana = grafana_mcp
        self._redis = redis
```

This makes testing possible and makes dependencies explicit.

-----

### Rule 7: Session State is Typed

Every Redis session key has a corresponding Pydantic model.
Never read or write raw dicts to session state.

```python
# BANNED
await redis.set(session_id, json.dumps({"error_rate": 0.5, "stuff": [...]}))

# REQUIRED
class MetricsFinding(BaseModel):
    cluster_id: str
    service: str
    error_rate_pct: float
    p99_latency_ms: float
    request_rate_rps: float
    anomaly_detected: bool
    anomaly_description: str
    baseline_comparison: dict[str, float]
    query_window_start: datetime
    query_window_end: datetime
    promql_queries_used: list[str]

await redis.set(
    f"session:{session_id}:metrics_findings",
    MetricsFinding(...).model_dump_json()
)
```

-----

### Rule 8: MCP Clients are Real Connections

Never abstract MCP clients to the point where they can silently fail.
Every MCP client must:

- Validate connection on initialization (call a health/list endpoint)
- Surface connection errors with actionable messages
- Log every tool call to Langfuse (tool name, inputs, latency, success/failure)

```python
class GrafanaMCPClient:
    def __init__(self, endpoint: str, token: str, langfuse: Langfuse):
        self._endpoint = endpoint
        self._token = token
        self._langfuse = langfuse
        self._session: aiohttp.ClientSession | None = None

    async def connect(self) -> None:
        """Initialize connection and validate Grafana is reachable."""
        self._session = aiohttp.ClientSession(
            headers={"Authorization": f"Bearer {self._token}"}
        )
        # Validate connection — fail fast, not silently
        try:
            datasources = await self.list_datasources()
            if not datasources:
                raise ConnectionError(
                    f"Grafana at {self._endpoint} returned no datasources. "
                    "Check service account permissions."
                )
        except aiohttp.ClientConnectorError as e:
            raise ConnectionError(
                f"Cannot reach Grafana MCP at {self._endpoint}: {e}. "
                "Verify GRAFANA_MCP_ENDPOINT and network policy."
            ) from e
```

-----

### Rule 9: Configuration is Validated at Startup

All environment variables and config values are validated before any agent runs.
Use Pydantic Settings. If a required value is missing, crash immediately with
a clear message — do not discover missing config mid-investigation.

```python
class AgentConfig(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8")

    grafana_mcp_endpoint: str
    redis_url: str
    qdrant_url: str
    jira_base_url: str
    jira_api_token: str
    slack_bot_token: str
    anthropic_api_key: str
    langfuse_secret_key: str
    langfuse_public_key: str
    langfuse_host: str

    @field_validator("grafana_mcp_endpoint")
    @classmethod
    def grafana_must_be_url(cls, v: str) -> str:
        if not v.startswith("http"):
            raise ValueError(f"GRAFANA_MCP_ENDPOINT must be a URL, got: {v}")
        return v
```

If any field is missing, Pydantic raises on import. The agent never starts.

-----

### Rule 10: Context Map Must Be Loaded Before Any Investigation

Every agent that touches a cluster MUST load the context map first.
This is not optional and not lazy-loaded.

```python
async def _load_context(self, session_id: str) -> InvestigationContext:
    """Load cluster context map and session state. Must be called first."""
    raw_map = await self._redis.get("context:map:global")
    if raw_map is None:
        raise RuntimeError(
            "Context map not found in Redis. "
            "Run 'scripts/seed_context_map.py' before starting agents."
        )
    context_map = ContextMap.model_validate_json(raw_map)

    raw_session = await self._redis.get(f"session:{session_id}")
    if raw_session is None:
        raise KeyError(f"Session '{session_id}' not found. TriageAgent must run first.")

    return InvestigationContext(
        map=context_map,
        session=InvestigationSession.model_validate_json(raw_session)
    )
```

-----

## Code Review Checklist

Before submitting any code, run through this list explicitly.
Do not assume. Check each item.

```
□ Every function has a type signature
□ Every function has a docstring
□ Every import is used
□ Every function is called from somewhere
□ Every class is instantiated somewhere
□ No TODO or FIXME comments exist
□ No mock/fake/placeholder data in production paths
□ No bare except: pass
□ No commented-out code blocks
□ All Redis writes use typed Pydantic models
□ All MCP clients validated on connection
□ Configuration validated at startup via Pydantic Settings
□ Context map loaded before any investigation begins
□ Error messages include actionable remediation steps
```

-----

## Project Structure

```
observability-agent/
├── CLAUDE.md                    ← this file
├── agents/
│   ├── triage.py                ← TriageAgent
│   ├── metrics.py               ← MetricsAgent
│   ├── logs.py                  ← LogsAgent
│   ├── k8s.py                   ← K8sAgent
│   ├── change_correlation.py    ← ChangeCorrelationAgent
│   ├── hypothesis.py            ← HypothesisAgent
│   ├── output.py                ← OutputAgent
│   ├── feedback.py              ← FeedbackAgent
│   ├── runbook.py               ← RunbookAgent
│   ├── judge.py                 ← JudgeAgent (different model)
│   ├── scheduler.py             ← SchedulerAgent (CronJob)
│   ├── false_positive.py        ← FalsePositiveTrackerAgent
│   └── staleness.py             ← StalenessMonitorAgent
├── mcps/
│   ├── grafana.py               ← GrafanaMCPClient (real connection)
│   ├── kubernetes.py            ← K8sMCPClient (real connection)
│   ├── jira.py                  ← JiraMCPClient (real connection)
│   ├── redis_client.py          ← RedisClient (real connection)
│   ├── qdrant_client.py         ← QdrantClient (real connection)
│   └── slack.py                 ← SlackMCPClient (real connection)
├── models/
│   ├── context_map.py           ← ContextMap, ClusterConfig, ServiceConfig
│   ├── session.py               ← InvestigationSession, all session state models
│   ├── findings.py              ← MetricsFinding, LogFinding, K8sFinding, etc.
│   ├── runbook.py               ← Runbook, JudgeScore
│   └── config.py                ← AgentConfig (Pydantic Settings)
├── scripts/
│   ├── seed_context_map.py      ← Seeds Redis with context-map.yaml
│   └── seed_baselines.py        ← Seeds per-cluster metric baselines from Mimir
├── k8s/
│   ├── deployments/             ← K8s manifests for each agent
│   ├── cronjobs/                ← SchedulerAgent, StalenessMonitor CronJobs
│   └── mcp-servers/             ← MCP server K8s manifests per cluster
├── tests/
│   ├── unit/                    ← Unit tests with mocks (mocks ONLY here)
│   └── integration/             ← Integration tests against real infra
└── context-map.yaml             ← Cluster topology, service catalog, baselines
```

-----

## What To Do When You Hit a Blocker

If you cannot implement something fully because:

- An MCP server isn't deployed yet
- A credential isn't available
- An API spec is unclear
- Infrastructure doesn't exist yet

**Do this:**

1. Stop writing code for that component
1. Tell the user exactly what is missing
1. Write a `docs/blocked/{component}.md` describing what is needed and why
1. Move on to a component that IS unblocked

**Do not:**

- Write a fake implementation that compiles but doesn't work
- Leave a TODO comment and continue
- Mock the missing piece silently

-----

## On Incremental Delivery

It is acceptable to build one agent before another.
It is NOT acceptable to build an agent incompletely.

If you are building MetricsAgent:

- MetricsAgent must be fully functional when you say it is done
- Its Grafana MCP client must be real
- Its Redis writes must be typed
- Its error handling must be complete

Do not deliver 10 half-built agents.
Deliver 1 fully-built agent, then build the next one.

-----

## Reminder: This Is Production Infrastructure

The agents in this system will:

- Query live Kubernetes clusters
- Read real metric and log data
- Create real Jira tickets
- Page real engineers
- Write runbooks that other agents will use to investigate incidents

There is no staging environment for "vibe-coded" code.
Every function you write will touch real systems.
Write accordingly.
