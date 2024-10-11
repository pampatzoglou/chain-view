graph TD
    A[Config] --> B[EndpointPool]
    B --> C[Endpoints]
    B --> D[CircuitBreaker]
    B --> E[RateLimiter]
    B --> F[JobQueue]
    F --> G[Workers]
    G --> H[HTTP Client]
    B --> I[Prometheus Metrics]
    J[Logger] --> B
    K[Context] --> B
