import * as child_process from 'child_process'
import { RemoteRepo } from '../../../shared/classes/remote'
import { ClonedRepository } from '../../../shared/classes/repository'
import { BundleServerWorldBase } from '../../../shared/support/world'
import { setWorldConstructor } from '@cucumber/cucumber'

export class IntegrationBundleServerWorld extends BundleServerWorldBase {
  remote: RemoteRepo | undefined
  local: ClonedRepository | undefined

  commandResult: child_process.SpawnSyncReturns<Buffer> | undefined

  runCommand(commandArgs: string): void {
    this.commandResult = child_process.spawnSync(`${this.parameters.bundleServerCommand} ${commandArgs}`, [], { shell: true })
  }

  runShell(command: string, ...args: string[]): child_process.SpawnSyncReturns<Buffer> {
    return child_process.spawnSync(command, args, { shell: true })
  }

  cloneRepository(): void {
    if (!this.remote) {
      throw new Error("Remote repository is not initialized")
    }

    const repoRoot = `${this.trashDirectory}/client`
    this.local = new ClonedRepository(this.remote, repoRoot)
  }
}

setWorldConstructor(IntegrationBundleServerWorld)
