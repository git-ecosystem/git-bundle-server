import { Given } from '@cucumber/cucumber'
import { IntegrationBundleServerWorld } from '../support/world'
import * as utils from '../../../shared/support/utils'
import { randomBytes } from 'crypto'

Given('the remote is cloned', async function (this: IntegrationBundleServerWorld) {
  this.cloneRepository()
})

Given('{int} commits are pushed to the remote branch {string}', async function (this: IntegrationBundleServerWorld, commitNum: number, branch: string) {
  if (this.local) {
    for (let i = 0; i < commitNum; i++) {
      utils.assertStatus(0, this.runShell(`echo ${randomBytes(16).toString('hex')} >${this.local.root}/README.md`))
      utils.assertStatus(0, utils.runGit("-C", this.local.root, "add", "README.md"))
      utils.assertStatus(0, utils.runGit("-C", this.local.root, "commit", "-m", `test ${i + 1}`))
    }
  } else {
    throw new Error("Local repo not initialized")
  }

  if (this.remote) {
    utils.assertStatus(0, utils.runGit("-C", this.local.root, "push", "origin", branch))
  } else {
    throw new Error("Remote repo not initialized")
  }
})
