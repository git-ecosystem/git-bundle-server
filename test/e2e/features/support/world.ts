import { setWorldConstructor, World, IWorldOptions } from '@cucumber/cucumber'
import * as utils from './utils'
import { BundleServer } from '../classes/bundleServer'

interface BundleServerParameters {
  bundleWebServerCommand: string
}

export class BundleServerWorld extends World<BundleServerParameters> {
  // Bundle server
  bundleServer: BundleServer

  constructor(options: IWorldOptions<BundleServerParameters>) {
    super(options)

    // Set up the bundle server
    this.bundleServer = new BundleServer(utils.absPath(this.parameters.bundleWebServerCommand))
  }

  cleanup(): void {
    this.bundleServer.cleanup()
  }
}

setWorldConstructor(BundleServerWorld)
