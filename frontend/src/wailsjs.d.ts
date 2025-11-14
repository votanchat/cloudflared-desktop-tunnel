// Wails Runtime Type Definitions
// This file provides TypeScript definitions for Wails runtime

declare global {
  interface Window {
    go: {
      app: {
        App: {
          StartTunnel(manualToken: string): Promise<void>;
          StopTunnel(): Promise<void>;
          GetTunnelStatus(): Promise<any>;
          GetConfig(): Promise<any>;
          UpdateConfig(config: any): Promise<void>;
          Greet(name: string): Promise<string>;
        };
      };
    };
    runtime: any;
  }
}

export {};
