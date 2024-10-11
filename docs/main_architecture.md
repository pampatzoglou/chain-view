graph TD
    A[Config] --> B[Main]
    B --> C[Logger]
    B --> D[Database]
    B --> E[Redis]
    B --> F[Endpoint Pools]
    B --> G[HTTP Server]
    G --> H[Health Check]
    G --> I[Startup Check]
    G --> J[Log Level Update]
    G --> K[Metrics]
    F --> L[Workers]
    B --> M[Graceful Shutdown]

graph TD
    A[Config] -->|Loads| B[Main]
    B -->|Initializes| C[Logger]
    B -->|Connects to| D[Database]
    B -->|Connects to| E[Redis]
    B -->|Creates| F[Endpoint Pools]
    F -->|Contains| F1[Chain 1]
    F -->|Contains| F2[Chain 2]
    F -->|Contains| F3[Chain n]
    F1 -->|Manages| EP1[Endpoints]
    F2 -->|Manages| EP2[Endpoints]
    F3 -->|Manages| EP3[Endpoints]
    B -->|Starts| G[HTTP Server]
    G -->|Handles| H[Health Check]
    G -->|Handles| I[Startup Check]
    G -->|Handles| J[Log Level Update]
    G -->|Exposes| K[Prometheus Metrics]
    F -->|Processed by| L[Workers]
    L -->|Use| M1[Circuit Breaker]
    L -->|Use| M2[Rate Limiter]
    B -->|Implements| N[Graceful Shutdown]
    O[OS Signals] -->|Triggers| N

sequenceDiagram
    participant M as Main
    participant C as Config
    participant L as Logger
    participant EP as Endpoint Pools
    participant W as Workers
    participant HS as HTTP Server
    participant S as Shutdown Signal

    M->>C: LoadConfig
    C-->>M: Return Config
    M->>L: Initialize Logger
    M->>EP: CreatePools
    EP-->>M: Return Pools
    M->>W: Start Workers (for each pool)
    M->>HS: Set up HTTP handlers
    M->>HS: Start HTTP Server
    
    par HTTP Server Running
        HS->>HS: Handle Requests
    and Workers Processing
        W->>W: Process Endpoints
    end

    S->>M: Receive Shutdown Signal
    M->>HS: Initiate Graceful Shutdown
    M->>W: Cancel Context (stop workers)
    M->>M: Wait for all goroutines to finish
    M->>M: Exit

sequenceDiagram
    participant M as Main
    participant C as Config
    participant L as Logger
    participant EP as Endpoint Pools
    participant W as Workers
    participant CB as Circuit Breaker
    participant RL as Rate Limiter
    participant HS as HTTP Server
    participant S as Shutdown Signal

    M->>C: LoadConfig
    C-->>M: Return Config
    M->>L: Initialize Logger
    M->>EP: CreatePools
    EP-->>M: Return Pools
    loop For each pool
        M->>W: Start Workers
        W->>CB: Initialize
        W->>RL: Initialize
    end
    M->>HS: Set up HTTP handlers
    M->>HS: Start HTTP Server
    
    par HTTP Server Running
        loop Handle Requests
            HS->>HS: Health Check
            HS->>HS: Startup Check
            HS->>HS: Log Level Update
            HS->>HS: Prometheus Metrics
        end
    and Workers Processing
        loop For each endpoint
            W->>CB: Check if allowed
            W->>RL: Check rate limit
            W->>EP: Fetch data from endpoint
            W->>L: Log result
            W->>CB: Update state
        end
    end

    S->>M: Receive Shutdown Signal
    M->>HS: Initiate Graceful Shutdown
    M->>W: Cancel Context (stop workers)
    M->>M: Wait for all goroutines to finish
    M->>M: Close database and Redis connections
    M->>M: Exit
