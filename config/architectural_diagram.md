classDiagram
    Config *-- ServerConfig
    Config*-- DatabaseConfig
    Config *-- RedisConfig
    Config*-- ChainConfig
    Config *-- GlobalSettings
    ServerConfig*-- LoggingConfig
    ChainConfig *-- EndpointConfig
    EndpointConfig*-- Duration
    GlobalSettings *-- Duration

    class Config {
        +ServerConfig Server
        +DatabaseConfig Database
        +RedisConfig Redis
        +ChainConfig[] Chains
        +GlobalSettings GlobalSettings
    }
    class ServerConfig {
        +int Port
        +LoggingConfig Logging
    }
    class LoggingConfig {
        +string Level
    }
    class DatabaseConfig {
        +string URL
    }
    class RedisConfig {
        +string URL
    }
    class ChainConfig {
        +int ChainID
        +string Network
        +EndpointConfig[] Endpoints
        +string PoolingStrategy
        +int RetryCount
        +Duration RetryBackoff
    }
    class EndpointConfig {
        +string Name
        +string URL
        +Duration Timeout
    }
    class Duration {
        +time.Duration Duration
        +UnmarshalYAML(unmarshal func(interface{}) error) error
    }
    class GlobalSettings {
        +Duration RequestTimeout
        +int MaxRetries
        +int MaxWorkers
        +Duration RetryBackoff
    }
