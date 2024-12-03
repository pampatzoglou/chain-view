sequenceDiagram
    participant A as Config
    participant LC as LoadConfig
    participant IO as ioutil
    participant YM as yaml.Unmarshal
    participant C as Config

    A->>LC: LoadConfig(filename)
    LC->>IO: ReadFile(filename)
    IO-->>LC: file data
    LC->>YM: Unmarshal(data, &config)
    YM->>C: Create Config struct
    YM->>C: Populate fields
    C-->>YM: Populated Config
    YM-->>LC: Unmarshaled Config
    LC-->>A: Return Config
