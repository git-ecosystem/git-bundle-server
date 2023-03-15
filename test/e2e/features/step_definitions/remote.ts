import { Given, } from '@cucumber/cucumber'
import { RemoteRepo } from '../classes/remote'
import { BundleServerWorld } from '../support/world'

/**
 * Steps relating to the setup of the remote repository users will clone from.
 */

Given('a remote repository {string}', async function (this: BundleServerWorld, url: string) {
  this.remote = new RemoteRepo(url)
})
