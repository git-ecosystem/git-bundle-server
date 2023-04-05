import * as child_process from 'child_process'
import { RemoteRepo } from './remote'
import * as utils from '../support/utils'

export class ClonedRepository {
  private initialized: boolean
  private root: string
  private remote: RemoteRepo | undefined

  cloneResult: child_process.SpawnSyncReturns<Buffer>
  cloneTimeMs: number

  constructor(remote: RemoteRepo, root: string, bundleUri?: string) {
    this.initialized = false
    this.remote = remote
    this.root = root

    // Clone the remote repository
    let args = ["clone"]
    if (bundleUri) {
      args.push(`--bundle-uri=${bundleUri}`)
    }
    args.push(this.remote.remoteUri, this.root)

    const timer = performance.now()
    this.cloneResult = child_process.spawnSync("git", args)
    this.cloneTimeMs = performance.now() - timer
    if (!this.cloneResult.error) {
      this.initialized = true
    }
  }

  runShell(command: string, ...args: string[]): child_process.SpawnSyncReturns<Buffer> {
    if (!this.initialized) {
      throw new Error("Repository is not initialized")
    }
    return child_process.spawnSync(command, args, { shell: true, cwd: this.root })
  }

  runGit(...args: string[]): child_process.SpawnSyncReturns<Buffer> {
    if (!this.initialized) {
      throw new Error("Repository is not initialized")
    }
    return utils.runGit("-C", this.root, ...args)
  }
}
