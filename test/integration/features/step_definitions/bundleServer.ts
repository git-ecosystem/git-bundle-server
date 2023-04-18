import * as assert from 'assert'
import { IntegrationBundleServerWorld } from '../support/world'
import { Given, Then } from '@cucumber/cucumber'
import * as utils from '../../../shared/support/utils'
import * as fs from 'fs'

Given('a bundle server repository is created at route {string} for the remote', async function (this: IntegrationBundleServerWorld, route: string) {
  if (!this.remote) {
    throw new Error("Remote has not been initialized")
  }
  this.bundleServer.init(this.remote, 'integration', route)
})

Given('no bundle server repository exists at route {string}', async function (this: IntegrationBundleServerWorld, route: string) {
  var repoPath = utils.repoRoot(route)
  if (fs.existsSync(repoPath)) {
    throw new Error(`Repo already exists at ${repoPath}`)
  }
})

Then('a bundle server repository exists at route {string}', async function (this: IntegrationBundleServerWorld, route: string) {
  var repoRoot = utils.repoRoot(route)
  assert.equal(fs.existsSync(repoRoot), true)
  assert.equal(fs.existsSync(`${repoRoot}/.git`), false)
  assert.equal(fs.existsSync(`${repoRoot}/HEAD`), true)
  assert.equal(fs.existsSync(`${repoRoot}/bundle-list.json`), true)

  // Set route for cleanup
  this.bundleServer.route = route
})

Then('the route configuration and repository data at {string} are removed', async function (this: IntegrationBundleServerWorld, route: string) {
  var repoRoot = utils.repoRoot(route)
  var routeData = fs.readFileSync(utils.routesPath())

  assert.equal(fs.existsSync(repoRoot), false)
  assert.equal(routeData.includes(route), false)

  // Reset route to be ignored in cleanup
  this.bundleServer.route = undefined
})

Then('the bundles are fetched and the bundle list is updated', async function (this: IntegrationBundleServerWorld) {
  assert.strictEqual(this.commandResult?.stdout.toString()
    .includes('Updating bundle list\n' +
              'Writing updated bundle list\n' +
              'Update complete'), true)

  if (this.bundleServer.initialBundleCount) {
    const currentBundleCount = this.bundleServer.getBundleCount()
    assert.strictEqual(currentBundleCount > this.bundleServer.initialBundleCount, true)
  } else {
    throw new Error("Bundle server not initialized")
  }
})

Then('the route is removed from the routes file', async function (this: IntegrationBundleServerWorld) {
  if (this.bundleServer.route) {
    var routesPath = utils.routesPath()
    var data = fs.readFileSync(routesPath);
    assert.strictEqual(data.includes(this.bundleServer.route), false)
  }
})

Then('the route exists in the routes file', async function (this: IntegrationBundleServerWorld) {
  if (this.bundleServer.route) {
    var routesPath = utils.routesPath()
    var data = fs.readFileSync(routesPath);
    assert.strictEqual(data.includes(this.bundleServer.route), true)
  } else {
    throw new Error("Route not set")
  }
})
