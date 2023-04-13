import { After, Given, Then } from "@cucumber/cucumber"
import { IntegrationBundleServerWorld } from "../support/world"
import { DaemonState } from '../support/daemonState'
import * as assert from 'assert'

Given('the daemon has not been started', async function (this: IntegrationBundleServerWorld) {
  var daemonState = this.getDaemonState()
  if (daemonState === DaemonState.Running) {
    this.runCommand('git-bundle-server web-server stop')
  }
})

Given('the daemon was started', async function (this: IntegrationBundleServerWorld) {
  var daemonState = this.getDaemonState()
  if (daemonState === DaemonState.NotRunning) {
    this.runCommand('git-bundle-server web-server start')
  }
})

Then('the daemon is running', async function (this: IntegrationBundleServerWorld) {
  var daemonStatus = this.getDaemonState()
  assert.strictEqual(daemonStatus, DaemonState.Running)
})

Then('the daemon is not running', async function (this: IntegrationBundleServerWorld) {
  var daemonStatus = this.getDaemonState()
  assert.strictEqual(daemonStatus, DaemonState.NotRunning)
})

// After({tags: '@daemon'}, async function (this: IntegrationBundleServerWorld) {
//   var daemonState = this.getDaemonState()
//   if (daemonState === DaemonState.Running) {
//     this.runCommand('git-bundle-server web-server stop')
//   }
// });
