import * as child_process from 'child_process'

export class BundleServer {
  private bundleWebServerCmd: string

  // Web server
  private webServerProcess: child_process.ChildProcess | undefined

  constructor(bundleWebServerCmd: string) {
    this.bundleWebServerCmd = bundleWebServerCmd
  }

  startWebServer(port: number): void {
    if (this.webServerProcess) {
      throw new Error("Tried to start web server, but web server is already running")
    }
    this.webServerProcess = child_process.spawn(this.bundleWebServerCmd, ["--port", String(port)])
  }

  cleanup(): void {
    if (this.webServerProcess) {
      const killed = this.webServerProcess.kill('SIGINT')
      if (!killed) {
        console.warn("Web server process was not successfully stopped")
      }
    }
  }
}
