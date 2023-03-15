import { BundleServerWorld, } from '../support/world'
import { Given } from '@cucumber/cucumber'

Given('the bundle web server was started at port {int}', async function (this: BundleServerWorld, port: number) {
  this.bundleServer.startWebServer(port)
})
