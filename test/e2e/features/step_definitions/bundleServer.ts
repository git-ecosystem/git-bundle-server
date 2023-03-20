import * as utils from '../support/utils'
import { BundleServerWorld, } from '../support/world'
import { Given } from '@cucumber/cucumber'

Given('the bundle web server was started at port {int}', async function (this: BundleServerWorld, port: number) {
  this.bundleServer.startWebServer(port)
})

Given('the bundle server has been initialized with the remote repo', async function (this: BundleServerWorld) {
  if (this.remote === undefined) {
    throw new Error("Remote repository is not initialized")
  }
  utils.assertStatus(0, this.bundleServer.init(this.remote))
})

Given('the bundle server was updated for the remote repo', async function (this: BundleServerWorld) {
  utils.assertStatus(0, this.bundleServer.update())
})
