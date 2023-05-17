import { Given } from '@cucumber/cucumber'
import { BundleServerWorldBase } from '../../support/world'

Given('the bundle web server was started at port {int}', async function (this: BundleServerWorldBase, port: number) {
  await this.bundleServer.startWebServer(port)
})
