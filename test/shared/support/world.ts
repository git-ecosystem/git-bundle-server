import { setWorldConstructor, World, IWorldOptions } from '@cucumber/cucumber'
import { randomUUID } from 'crypto'
import { RemoteRepo } from '../classes/remote'
import * as utils from './utils'
import * as fs from 'fs'
import * as path from 'path'
import { BundleServer } from '../classes/bundleServer'

export interface BundleServerParameters {
  bundleServerCommand: string
  bundleWebServerCommand: string
  trashDirectoryBase: string
}

export class BundleServerWorldBase extends World<BundleServerParameters> {
  // Internal variables
  trashDirectory: string

  // Bundle server
  bundleServer: BundleServer
  remote: RemoteRepo | undefined

  constructor(options: IWorldOptions<BundleServerParameters>) {
    super(options)

    // Set up the trash directory
    this.trashDirectory = path.join(utils.absPath(this.parameters.trashDirectoryBase), randomUUID())
    fs.mkdirSync(this.trashDirectory, { recursive: true });

    // Set up the bundle server
    this.bundleServer = new BundleServer(utils.absPath(this.parameters.bundleServerCommand),
      utils.absPath(this.parameters.bundleWebServerCommand))
  }

  cleanup(): void {
    this.bundleServer.cleanup()

    // Delete the trash directory
    fs.rmSync(this.trashDirectory, { recursive: true })
  }
}
