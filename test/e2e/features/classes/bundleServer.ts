import { randomBytes } from 'crypto'
import * as child_process from 'child_process'
import { RemoteRepo } from './remote'

export class BundleServer {
  private bundleServerCmd: string
  private bundleWebServerCmd: string

  // Web server
  private webServerProcess: child_process.ChildProcess | undefined
  private bundleUriBase: string | undefined

  // Remote repo info (for now, only support one per test)
  private route: string | undefined

  constructor(bundleServerCmd: string, bundleWebServerCmd: string) {
    this.bundleServerCmd = bundleServerCmd
    this.bundleWebServerCmd = bundleWebServerCmd
  }

  startWebServer(port: number): void {
    if (this.webServerProcess) {
      throw new Error("Tried to start web server, but web server is already running")
    }
    this.webServerProcess = child_process.spawn(this.bundleWebServerCmd, ["--port", String(port)])
    this.bundleUriBase = `http://localhost:${port}/`
  }

  init(remote: RemoteRepo): child_process.SpawnSyncReturns<Buffer> {
    this.route = `e2e/${randomBytes(8).toString('hex')}`
    return child_process.spawnSync(this.bundleServerCmd, ["init", remote.remoteUri, this.route])
  }

  bundleUri(): string {
    if (!this.webServerProcess) {
      throw new Error("Tried to get bundle URI before starting the web server")
    }
    if (!this.route) {
      throw new Error("Tried to get bundle URI before running 'init'")
    }

    return this.bundleUriBase + this.route
  }

  cleanup(): void {
    if (this.webServerProcess) {
      const killed = this.webServerProcess.kill('SIGINT')
      if (!killed) {
        console.warn("Web server process was not successfully stopped")
      }
    }

    // Delete the added route
    if (this.route) {
      child_process.spawnSync(this.bundleServerCmd, ["delete", this.route])
    }
  }
}
