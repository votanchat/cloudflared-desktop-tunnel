# Cloudflared Desktop Tunnel - Flow Charts

## 1. Application Startup Flow

```
┌─────────────────────────────────────┐
│     User Launches Application       │
└────────────┬────────────────────────┘
             │
             ▼
┌─────────────────────────────────────┐
│  Wails Runtime Initializes          │
│  (main.go → wails.Run())            │
└────────────┬────────────────────────┘
             │
             ▼
┌─────────────────────────────────────┐
│  App.Startup(ctx) Called            │
│  - Save context                     │
│  - Load configuration               │
│  - Initialize BackendClient         │
│  - Initialize TunnelManager         │
└────────────┬────────────────────────┘
             │
             ▼
         ┌───┴───┐
         │       │
         ▼       ▼
   ┌──────────┐ ┌──────────────────┐
   │AutoStart?│ │Load Config Error?│
   └────┬─────┘ └────────┬─────────┘
        │                │
    YES │            YES │
        ▼                ▼
   ┌─────────┐    ┌──────────────┐
   │Start    │    │Use Default   │
   │Tunnel   │    │Config        │
   │(async)  │    └──────────────┘
   └─────────┘
        │
        ▼
┌─────────────────────────────────────┐
│  App.DomReady(ctx) Called           │
│  - Frontend is ready                │
│  - Display UI                       │
└─────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────┐
│  Application Running                │
│  - Listening to user interactions   │
│  - WebSocket connected (if backend) │
│  - Token refresh loop active        │
└─────────────────────────────────────┘
```

---

## 2. Tunnel Start Flow - Manual Token Path

```
┌──────────────────────────────┐
│  User clicks Start Tunnel    │
└────────────┬─────────────────┘
             │
             ▼
    ┌────────────────┐
    │Manual Token?   │
    │Provided?       │
    └────┬───────┬───┘
         │       │
       YES│       │NO
         │       │
         ▼       ▼
    ┌────────┐  ┌─────────────────────┐
    │Use     │  │Fetch from Backend   │
    │Manual  │  │GET /api/token       │
    │Token   │  │                     │
    └────┬───┘  └────────┬────────────┘
         │               │
         ▼               ▼
    ┌─────────────────────────────────┐
    │  TunnelManager.Start(token)     │
    │  Called                         │
    └────────────┬────────────────────┘
                 │
                 ▼
    ┌─────────────────────────────────┐
    │  Tunnel Already Running?        │
    │  (Check IsRunning())            │
    └────────┬────────────┬───────────┘
             │            │
          NO │        YES │
             │            ▼
             │      ┌──────────────┐
             │      │Return Error  │
             │      │"Already      │
             │      │Running"      │
             │      └──────────────┘
             ▼
    ┌─────────────────────────────────┐
    │  Ensure Binary Downloaded       │
    │  ensureBinary()                 │
    │  - Check cache                  │
    │  - Download if needed           │
    │  - Verify integrity             │
    └────────────┬────────────────────┘
                 │
                 ├──────────────┬─────────────┐
                 │              │             │
            Cache Hit?      Download Error?   │
                 │              │             │
                 ▼              ▼             │
          ┌──────────┐   ┌─────────────┐    │
          │Use Cached│   │Return Error │    │
          │Binary    │   │"Failed to   │    │
          │          │   │Download"    │    │
          └────┬─────┘   └─────────────┘    │
               │                            │
               └──────────────┬─────────────┘
                              │
                              ▼
    ┌─────────────────────────────────┐
    │  Check if Routes Configured     │
    │  (config.Routes != empty?)      │
    └────────┬────────────┬───────────┘
             │            │
          NO │        YES │
             │            │
             ▼            ▼
    ┌──────────────┐  ┌──────────────────┐
    │Use Token     │  │Generate Config   │
    │Mode Only     │  │File with Routes  │
    │(Dashboard    │  │                  │
    │managed)      │  └────┬─────────────┘
    └────┬─────────┘       │
         │                 ▼
         │       ┌──────────────────────┐
         │       │Write YAML Config     │
         │       │- Ingress routes      │
         │       │- Catch-all 404       │
         │       │- Save to config dir  │
         │       └────┬─────────────────┘
         │            │
         ▼            ▼
    ┌─────────────────────────────────┐
    │  Prepare Command                │
    │  exec.Command(binary, args...)  │
    │  With token and optional config │
    └────────────┬────────────────────┘
                 │
                 ▼
    ┌─────────────────────────────────┐
    │  Create Pipes for Output        │
    │  - StdoutPipe()                 │
    │  - StderrPipe()                 │
    └────────────┬────────────────────┘
                 │
                 ▼
    ┌─────────────────────────────────┐
    │  Start Process                  │
    │  cmd.Start()                    │
    │  Set running = true             │
    └────────────┬────────────────────┘
                 │
                 ▼
    ┌─────────────────────────────────┐
    │  Launch Goroutines              │
    │  - readLogs(stdout, "stdout")   │
    │  - readLogs(stderr, "stderr")   │
    │  - monitorProcess()             │
    └────────────┬────────────────────┘
                 │
                 ▼
    ┌─────────────────────────────────┐
    │  Return Success                 │
    │  Tunnel Running                 │
    │  Logs streaming to UI           │
    └─────────────────────────────────┘
```

---

## 3. Tunnel Start Flow - Backend Token Path

```
┌─────────────────────────────────────┐
│  User clicks Start Tunnel           │
│  (No Manual Token provided)         │
└────────────┬────────────────────────┘
             │
             ▼
┌─────────────────────────────────────┐
│  BackendClient.FetchToken()         │
│  Called                             │
└────────────┬────────────────────────┘
             │
             ▼
┌─────────────────────────────────────┐
│  Make HTTP GET Request              │
│  GET {baseURL}/api/token            │
│  Timeout: 30s                       │
└────────────┬────────────────────────┘
             │
        ┌────┴────┐
        │          │
        ▼          ▼
   ┌────────┐  ┌──────────┐
   │Success?│  │Network   │
   │(200 OK)│  │Error?    │
   └───┬────┘  └────┬─────┘
       │            │
      YES          │
       │            ▼
       │       ┌──────────────┐
       │       │Return Error  │
       │       │"Failed to    │
       │       │fetch token"  │
       │       └──────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Parse JSON Response                │
│  {                                  │
│    "token": "...",                  │
│    "expiresAt": "2025-11-15..."     │
│  }                                  │
└────────────┬────────────────────────┘
             │
             ▼
┌─────────────────────────────────────┐
│  Store token in BackendClient       │
│  Return token to caller             │
└────────────┬────────────────────────┘
             │
             ▼
┌─────────────────────────────────────┐
│  TunnelManager.Start(token)         │
│  Continue with normal tunnel start  │
│  (See Tunnel Start Flow)            │
└─────────────────────────────────────┘
```

---

## 4. Binary Download Flow

```
┌──────────────────────────────────────────┐
│  TunnelManager.ensureBinary() Called     │
│  Check if binary is available            │
└───────────────┬──────────────────────────┘
                │
                ▼
┌──────────────────────────────────────────┐
│  Cached Binary Path Exists?              │
│  (tm.binaryPath != "")                   │
└────────────┬──────────────────┬──────────┘
             │                  │
           YES│                 │NO
             │                  │
             ▼                  ▼
    ┌─────────────────┐  ┌──────────────────┐
    │Stat File        │  │GetCacheDir()     │
    │- Check exists   │  │- Platform-specific
    │- Get size       │  │- Create if needed
    └────┬────────────┘  └────┬─────────────┘
         │                    │
         ▼                    │
    ┌──────────┐             │
    │Exists?   │             │
    └───┬──┬───┘             │
        │ │                  │
      YES│ │NO               │
        │ ▼                  │
        │ ┌──────────┐       │
        │ │Re-download
        │ │Binary    │       │
        │ └────┬─────┘       │
        │      │             │
        ▼      ▼             │
    ┌──────────────────┐     │
    │Verify Binary     │     │
    │- Size >= 10MB    │     │
    │- Executable perms
    └────┬─────────────┘     │
         │                   │
         ├──────┬────────┐   │
         │      │        │   │
       OK│  INVALID│ CACHE MISS
         │      │        │   │
         ▼      ▼        │   │
    ┌─────────────────┐  │   │
    │Return Cached    │  │   │
    │Binary Path      │  │   │
    └─────────────────┘  │   │
                         │   │
                         ▼   ▼
            ┌──────────────────────────┐
            │  getLatestVersion()      │
            │  Fetch GitHub API        │
            │  GET .../releases/latest │
            └────────────┬─────────────┘
                         │
                         ▼
            ┌──────────────────────────┐
            │  Parse Release Tag       │
            │  Extract version number  │
            └────────────┬─────────────┘
                         │
                         ▼
            ┌──────────────────────────┐
            │  downloadBinary(version) │
            │  Build download URL      │
            │  based on GOOS/GOARCH    │
            └────────────┬─────────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
     Windows          macOS           Linux
         │               │               │
         ▼               ▼               ▼
    ┌────────┐  ┌────────────┐   ┌─────────┐
    │.exe    │  │.tgz archive│   │Binary   │
    │Direct  │  │Extract tar │   │Direct   │
    │Download│  │+ gzip      │   │Download │
    └────┬───┘  └────┬───────┘   └────┬────┘
         │           │                │
         └───────┬───┴────────────┬────┘
                 │                │
                 ▼                ▼
    ┌──────────────────────────────────┐
    │  Download from GitHub Release    │
    │  (5 minute timeout)              │
    │  Check HTTP Status 200 OK        │
    └────────────┬─────────────────────┘
                 │
                 ├─────┬──────┐
                 │     │      │
            Success   │    Error
                 │     │      │
                 ▼     ▼      ▼
         ┌──────────┐  ┌──────────┐
         │Write to  │  │Return    │
         │Output    │  │Error     │
         │Path      │  │"Download │
         │          │  │Failed"   │
         └────┬─────┘  └──────────┘
              │
              ▼
    ┌──────────────────────────────────┐
    │  Set Executable Permissions      │
    │  (Unix: chmod +x)                │
    │  (Windows: skip)                 │
    └────────────┬─────────────────────┘
                 │
                 ▼
    ┌──────────────────────────────────┐
    │  Verify Downloaded Binary        │
    │  - Check size                    │
    │  - Check permissions             │
    │  - Try to execute if needed      │
    └────────────┬─────────────────────┘
                 │
         ┌───────┴────────┐
         │                │
      Valid            Invalid
         │                │
         ▼                ▼
    ┌────────┐    ┌──────────────┐
    │Cache   │    │Delete file   │
    │Binary  │    │Return error  │
    │Return  │    │"Binary not   │
    │Path    │    │valid"        │
    └────────┘    └──────────────┘
```

---

## 5. Token Management Flow

```
                    ┌──────────────────────────────┐
                    │   Application Running        │
                    └────────────┬─────────────────┘
                                 │
                    ┌────────────┴──────────────┐
                    │                           │
                    ▼                           ▼
        ┌─────────────────────┐    ┌──────────────────────┐
        │User Action: Start   │    │Backend Token Refresh │
        │Tunnel              │    │Loop Active (5 min)   │
        └────────┬────────────┘    └──────────┬───────────┘
                 │                            │
                 ▼                            ▼
        ┌─────────────────────┐    ┌──────────────────────┐
        │Manual Token Input?  │    │Call FetchToken()     │
        │                     │    │GET /api/token        │
        └────┬────────┬───────┘    └──────────┬───────────┘
             │        │                       │
           YES│       │NO                     ▼
             │        │            ┌──────────────────────┐
             │        │            │Store new token       │
             │        │            │Update expiryAt       │
             │        │            │Log refresh           │
             ▼        ▼            └──────────────────────┘
    ┌────────────────────────────┐
    │Use Manual Token            │
    │(Skip Backend Call)         │
    │Token = manual input        │
    └─────────┬──────────────────┘
              │
              ▼
    ┌────────────────────────────┐
    │Start Tunnel with Token     │
    │TunnelManager.Start(token)  │
    └─────────┬──────────────────┘
              │
              ▼
    ┌────────────────────────────┐
    │Execute cloudflared process │
    │Pass token via --token flag │
    │or in config file (routes)  │
    └─────────┬──────────────────┘
              │
              ▼
    ┌────────────────────────────┐
    │Tunnel Active               │
    │Streaming logs              │
    │Status: running             │
    └────────────────────────────┘
```

---

## 6. Backend Communication Flow

```
┌────────────────────────────────────────────┐
│  App Running with Backend Configured       │
└────────────┬─────────────────────────────────┘
             │
    ┌────────┴──────────┬──────────────┐
    │                   │              │
    ▼                   ▼              ▼
┌──────────┐   ┌────────────────┐  ┌──────────────┐
│WebSocket │   │Token Refresh   │  │Status Report │
│Commands  │   │Loop (5 min)    │  │(on demand)   │
└────┬─────┘   └────────┬───────┘  └──────┬───────┘
     │                  │                  │
     ▼                  ▼                  ▼
┌────────────────────────────────────────────┐
│  BackendClient Methods                     │
└────────────┬─────────────────────────────────┘
             │
    ┌────────┼──────────────┐
    │        │              │
    ▼        ▼              ▼
┌─────────────────┐  ┌─────────────┐  ┌──────────────────┐
│connectWebSocket │  │FetchToken() │  │ReportStatus()    │
│(ctx)            │  │(Periodic)   │  │(on Demand)       │
└────────┬────────┘  └──────┬──────┘  └────────┬─────────┘
         │                  │                   │
         ▼                  ▼                   ▼
┌─────────────────────────┐   ┌──────────────┐
│Connect WS:              │   │HTTP GET      │
│ws/wss://base/api/...    │   │/api/token    │
└────────┬────────────────┘   └──────┬───────┘
         │                           │
         ▼                           ▼
    ┌──────────────┐        ┌────────────────┐
    │Connected?   │        │Got Token?      │
    │Read messages│        │Update expiryAt │
    └─────┬───────┘        └────────┬───────┘
          │                         │
          ▼                         ▼
    ┌──────────────────────────────────────┐
    │New Command Received                  │
    │{type, payload}                       │
    └────────┬─────────────────────────────┘
             │
    ┌────────┴────────┬──────────────┬──────┐
    │                 │              │      │
    ▼                 ▼              ▼      ▼
┌────────┐    ┌──────────┐  ┌────────┐  ┌────┐
│update  │    │restart   │  │patch   │  │stop│
└──┬─────┘    └────┬─────┘  └───┬────┘  └─┬──┘
   │               │            │        │
   ▼               ▼            ▼        ▼
┌─────────────────────────────────────────┐
│processCommands() Goroutine               │
│- Switch on command type                 │
│- Log command received                   │
│- Execute action                         │
│- (Full implementation in TODO items)    │
└─────────────────────────────────────────┘
```

---

## 7. Graceful Shutdown Flow

```
┌──────────────────────────────────┐
│  User Closes Application         │
│  OR System Shutdown Initiated    │
└──────────────┬───────────────────┘
               │
               ▼
┌──────────────────────────────────┐
│  Wails Triggers Shutdown Hook    │
│  App.Shutdown(ctx) Called        │
└──────────────┬───────────────────┘
               │
    ┌──────────┴──────────┐
    │                     │
    ▼                     ▼
┌────────────────┐  ┌───────────────────┐
│Stop Tunnel if  │  │Stop Backend Client│
│Running         │  │                   │
└────┬───────────┘  └────────┬──────────┘
     │                       │
     ▼                       ▼
┌────────────────┐  ┌───────────────────┐
│Check IsRunning │  │Close WebSocket    │
└────┬───────────┘  │Close HTTP client  │
     │              │Stop refresh loop  │
     ├──────────┐   └────────┬──────────┘
     │          │            │
   YES│        NO│            │
     │          │            │
     ▼          ▼            ▼
┌──────────────────────────────────┐
│Kill cloudflared Process          │
│cmd.Process.Kill()                │
│Wait for exit (cmd.Wait())        │
└──────────────┬───────────────────┘
               │
               ▼
┌──────────────────────────────────┐
│Clean Up Cached Binary            │
│(Optional - can be reused)        │
│Cleanup() → os.Remove(binaryPath) │
└──────────────┬───────────────────┘
               │
               ▼
┌──────────────────────────────────┐
│Save Configuration                │
│Config.Save() → JSON to file      │
│Persist user settings             │
└──────────────┬───────────────────┘
               │
               ▼
┌──────────────────────────────────┐
│Log Shutdown                      │
│Close all resources               │
│Exit application gracefully       │
└──────────────────────────────────┘
```

---

## 8. Configuration Flow

```
┌────────────────────────────────────┐
│  Application Startup               │
│  OR User Opens Settings            │
└──────────────┬─────────────────────┘
               │
               ▼
┌────────────────────────────────────┐
│  LoadConfig() Called               │
│  - Get config path (platform-spec) │
└──────────────┬─────────────────────┘
               │
               ▼
┌────────────────────────────────────┐
│  Config File Exists?               │
│  getConfigPath() returns path      │
└────────┬──────────────────┬────────┘
         │                  │
       YES│                 │NO
         │                  │
         ▼                  ▼
┌──────────────────┐  ┌────────────────┐
│Read File         │  │Use Default     │
│os.ReadFile()     │  │Config          │
└────┬─────────────┘  │DefaultConfig() │
     │                └────┬───────────┘
     ▼                     │
┌──────────────────┐       │
│Parse JSON        │       │
│json.Unmarshal()  │       │
└────┬─────────────┘       │
     │                     │
     ├──────┬──────┐       │
     │      │      │       │
  JSON   Parse  │ Error    │
  Valid  Error  │          │
     │      │      │       │
     ▼      ▼      ▼       │
     │   ┌─────────┐        │
     │   │Return   │        │
     │   │Error    │        │
     │   └─────────┘        │
     │                     │
     └────────┬────────────┘
              │
              ▼
┌────────────────────────────────────┐
│  Config Object Created             │
│  {                                 │
│    BackendURL,                     │
│    TunnelName,                     │
│    AutoStart,                      │
│    Routes: []                      │
│  }                                 │
└──────────────┬─────────────────────┘
               │
    ┌──────────┴──────────┐
    │                     │
    ▼                     ▼
┌────────────────┐  ┌──────────────────┐
│App Startup     │  │User Updates      │
│Use Config      │  │Settings in UI    │
│AutoStart?      │  └────────┬─────────┘
└────┬───────────┘           │
     │                       ▼
     │              ┌──────────────────┐
     │              │UpdateConfig()    │
     │              │Called            │
     │              └────────┬─────────┘
     │                       │
     │                       ▼
     │              ┌──────────────────┐
     │              │Update fields     │
     │              │BackendURL        │
     │              │TunnelName        │
     │              │AutoStart         │
     │              │Etc...            │
     │              └────────┬─────────┘
     │                       │
     │                       ▼
     │              ┌──────────────────┐
     │              │Save Config File  │
     │              │Config.Save()     │
     │              │JSON to disk      │
     │              └────────┬─────────┘
     │                       │
     └───────────┬───────────┘
                 │
                 ▼
┌────────────────────────────────────┐
│  Config Persisted                  │
│  Location:                         │
│  - Windows: %APPDATA%\app\         │
│  - macOS: ~/Library/App Support/   │
│  - Linux: ~/.config/app/           │
└────────────────────────────────────┘
```

---

## 9. Log Management Flow

```
┌─────────────────────────────────────────┐
│  cloudflared Process Running            │
│  Outputting logs to stdout/stderr       │
└──────────────┬────────────────────────────┘
               │
    ┌──────────┴──────────┐
    │                     │
    ▼                     ▼
┌─────────────┐  ┌────────────────┐
│StdoutPipe   │  │StderrPipe      │
│Goroutine    │  │Goroutine       │
└─────┬───────┘  └────────┬───────┘
      │                   │
      ▼                   ▼
┌───────────────────────────────────┐
│readLogs(pipe, source) Goroutine   │
│- Buffer: 4KB                      │
│- Read from pipe in loop           │
│- Handle partial reads             │
└──────────────┬────────────────────┘
               │
               ▼
┌───────────────────────────────────┐
│Parse log line from buffer         │
│Convert bytes to string            │
│Trim/format as needed              │
└──────────────┬────────────────────┘
               │
               ▼
┌───────────────────────────────────┐
│Log to Console                     │
│log.Printf("[%s] %s", source, line)
└──────────────┬────────────────────┘
               │
               ▼
┌───────────────────────────────────┐
│Store in TunnelManager.logs        │
│- Acquire mutex lock               │
│- Append to logs slice             │
│- Keep only last 100 lines         │
│  (using copy + reslice)           │
│- Release mutex                    │
└──────────────┬────────────────────┘
               │
        ┌──────┴──────┐
        │             │
   EOF? │        Keep reading
        │             │
        ▼             ▼
   ┌────────┐    ┌─────────┐
   │Close   │    │Loop back│
   │Pipe    │    │to Read  │
   └────────┘    └─────────┘
        │
        ▼
┌──────────────────────────────┐
│GetLogs() Called              │
│(From Frontend or API)        │
│- Acquire read lock           │
│- Return copy of logs slice   │
│- Release lock                │
│- Send to frontend via Wails  │
└──────────────────────────────┘
        │
        ▼
┌──────────────────────────────┐
│Display Logs in UI            │
│- Status panel                │
│- Logs viewer component       │
│- Real-time updates           │
└──────────────────────────────┘
```

---

## 10. Error Handling Flow

```
┌─────────────────────────────────┐
│  Operation Initiated            │
│  (Start, Download, etc.)        │
└──────────────┬──────────────────┘
               │
               ▼
┌─────────────────────────────────┐
│  Error Occurs                   │
│  (network, file, permission...)  │
└──────────────┬──────────────────┘
               │
    ┌──────────┴──────────────────┐
    │                             │
    ▼                             ▼
┌──────────────────┐    ┌─────────────────┐
│Wrapped Error     │    │Logged with      │
│using %w          │    │log.Printf()     │
│fmt.Errorf()      │    │Full stack trace │
└────────┬─────────┘    └────────┬────────┘
         │                       │
         ▼                       ▼
┌──────────────────────────────────────┐
│Return Error to Caller                │
│- Propagate up call stack             │
│- Frontend receives error message     │
└──────────────┬──────────────────────┘
               │
    ┌──────────┴─────────┐
    │                    │
    ▼                    ▼
┌──────────────┐  ┌──────────────────┐
│Critical      │  │Non-critical      │
│Error?        │  │Error?            │
│(no binaries) │  │(backend offline) │
└────┬─────────┘  └────────┬─────────┘
     │                     │
     ▼                     ▼
┌──────────────┐  ┌──────────────────┐
│Stop Tunnel   │  │Fallback Option?  │
│Don't Retry   │  │(Manual Token)    │
│Log Critical  │  │Suggest to user   │
└──────────────┘  └──────────────────┘
```

---

## Flow Legend

```
┌─────┐
│ Box │  = Process/Action
└─────┘

  │     = Flow direction
  ▼     = Next step

  ├──┐  = Conditional/Decision
  │  │
  ▼  ▼  = Multiple paths

  ↔   = Bidirectional communication
```

---

## Component Interaction Sequence

```
Frontend          Wails           App             TunnelManager    Backend    GitHub
  │                 │              │                  │              │          │
  │ Click Start     │              │                  │              │          │
  ├────────────────→│              │                  │              │          │
  │                 │ StartTunnel()│                  │              │          │
  │                 ├─────────────→│                  │              │          │
  │                 │              │ FetchToken()    │              │          │
  │                 │              ├─────────────────────────────────→          │
  │                 │              │                  │              │ Response │
  │                 │              │←─────────────────────────────────          │
  │                 │              │ Start(token)    │              │          │
  │                 │              ├─────────────────→              │          │
  │                 │              │                  │ ensureBinary()         │
  │                 │              │                  ├──────────────────────→│
  │                 │              │                  │              │ Download  │
  │                 │              │                  │←──────────────────────│
  │                 │              │                  │ Start process         │
  │                 │              │                  ├──→ (cloudflared)      │
  │                 │              │                  │                       │
  │ Status Update   │ GetLogs()    │                  │                       │
  │←────────────────┼─────────────→│  GetLogs()      │                       │
  │                 │              ├──────────────────→                       │
  │                 │              │←──────────────────                       │
  │                 │←─────────────┤                  │                       │
  │←────────────────┤              │                  │                       │
```

