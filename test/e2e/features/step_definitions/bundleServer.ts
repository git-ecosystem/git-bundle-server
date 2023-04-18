import * as utils from '../../../shared/support/utils'
import { EndToEndBundleServerWorld } from '../support/world'
import { Given } from '@cucumber/cucumber'

Given('the bundle server has been initialized with the remote repo', async function (this: EndToEndBundleServerWorld) {
  if (this.remote === undefined) {
    throw new Error("Remote repository is not initialized")
  }
  utils.assertStatus(0, this.bundleServer.init(this.remote, 'e2e'))
})

Given('the bundle server was updated for the remote repo', async function (this: EndToEndBundleServerWorld) {
  utils.assertStatus(0, this.bundleServer.update())
})
