sequenceDiagram
    participant M as endpoints.go
    participant EP as EndpointPool
    participant CB as CircuitBreaker
    participant RL as RateLimiter
    participant JQ as JobQueue
    participant W as Worker
    participant HC as HTTP Client
    participant E as Endpoint

    M->>EP: StartWorkers
    M->>EP: ProcessEndpoints
    loop Every second
        EP->>EP: GetNextEndpoint
        EP->>JQ: Send Job
    end
    W->>JQ: Receive Job
    W->>RL: Wait (rate limit check)
    W->>CB: Allow (circuit breaker check)
    W->>HC: fetchData
    HC->>E: HTTP GET
    E-->>HC: Response
    HC-->>W: Response
    W->>EP: Update metrics
    W->>CB: RecordSuccess/RecordFailure
