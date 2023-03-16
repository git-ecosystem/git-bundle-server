import { Given, } from '@cucumber/cucumber'
import { RemoteRepo } from '../classes/remote'
import { BundleServerWorld } from '../support/world'
import * as path from 'path'

/**
 * Steps relating to the setup of the remote repository users will clone from.
 */

Given('a remote repository {string}', async function (this: BundleServerWorld, url: string) {
  this.remote = new RemoteRepo(false, url)
})

Given('a new remote repository with main branch {string}', async function (this: BundleServerWorld, mainBranch: string) {
  this.remote = new RemoteRepo(true, path.join(this.trashDirectory, "server"), mainBranch)
})
