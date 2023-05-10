import { randomBytes } from 'crypto'
import * as child_process from 'child_process'
import { RemoteRepo } from './remote'
import * as fs from 'fs'
import * as utils from '../support/utils'

export class BundleServer {
  private bundleServerCmd: string
  private bundleWebServerCmd: string

  // Web server
  private webServerProcess: child_process.ChildProcess | undefined
  private bundleUriBase: string | undefined

  // Remote repo info (for now, only support one per test)
  route: string | undefined
  initialBundleCount: number | undefined

  constructor(bundleServerCmd: string, bundleWebServerCmd: string) {
    this.bundleServerCmd = bundleServerCmd
    this.bundleWebServerCmd = bundleWebServerCmd
  }

  async startWebServer(port: number): Promise<void> {
    if (this.webServerProcess) {
      throw new Error("Tried to start web server, but web server is already running")
    }
    const webServerProcess = child_process.spawn(this.bundleWebServerCmd, ["--port", String(port)])
    this.webServerProcess = webServerProcess
    this.bundleUriBase = `http://localhost:${port}/`

    // Now, ensure the server is running
    var timer: NodeJS.Timeout | undefined
    var ok: boolean | undefined

    await Promise.race([
      new Promise<void>(resolve => webServerProcess.stdout.on('data', (data: string) => {
        if (data.includes("Server is running at address") && ok === undefined) {
          ok = true
          resolve() // server is running
        }
      })),
      new Promise<void>(resolve => webServerProcess.on('close', () => {
        ok = false
        resolve() // program failed to start/exited early
      })),
      new Promise(resolve => timer = setTimeout(resolve, 2000)) // fallback timeout
    ])

    // If it's still running, clear the timeout so it doesn't delay shutdown
    clearTimeout(timer)

    if (!ok) {
      throw new Error('Failed to start web server')
    }
  }

  init(remote: RemoteRepo, routePrefix: string, route: string = ""): child_process.SpawnSyncReturns<Buffer> {
    if (route === "") {
      route = `${routePrefix}/${randomBytes(8).toString('hex')}`
    }
    this.route = route

    const repoPath = utils.repoRoot(route)
    if (fs.existsSync(repoPath)) {
      throw new Error("Bundle server repository already exists")
    }

    const result = child_process.spawnSync(this.bundleServerCmd, ["init", remote.remoteUri, this.route])
    this.initialBundleCount = this.getBundleCount()

    return result
  }

  update(): child_process.SpawnSyncReturns<Buffer> {
    if (!this.route) {
      throw new Error("Tried to update server before running 'init'")
    }
    return child_process.spawnSync(this.bundleServerCmd, ["update", this.route])
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

  getBundleCount(): number {
    if (!this.route) {
      throw new Error("Route is not defined")
    }

    var matches: string[] = [];
    const files = fs.readdirSync(`${utils.wwwPath()}/${this.route}`);

    for (const file of files) {
      if (file.endsWith('.bundle')) {
        matches.push(file);
      }
    }

    return matches.length;
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
