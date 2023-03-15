import * as assert from 'assert'
import * as utils from '../support/utils'
import { BundleServerWorld, User } from '../support/world'
import { Given, When, Then } from '@cucumber/cucumber'

/**
 * Steps relating to the repository clones that users work with directly. Since
 * these steps represent the actions a user takes, the majority of end-to-end
 * test steps will live here.
 */

When('I clone from the remote repo with a bundle URI', async function (this: BundleServerWorld) {
  this.cloneRepositoryFor(User.Me, this.bundleServer.bundleUri())
})

Then('bundles are downloaded and used', async function (this: BundleServerWorld) {
  const clonedRepo = this.getRepo(User.Me)

  // Verify the clone executed as-expected
  utils.assertStatus(0, clonedRepo.cloneResult, "git clone failed")

  // Ensure warning wasn't thrown
  clonedRepo.cloneResult.stderr.toString().split("\n").forEach(function (line) {
    if (line.startsWith("warning: failed to download bundle from URI")) {
      assert.fail(line)
    }
  })

  // Make sure the config is set up properly
  let result = clonedRepo.runGit("config", "--get", "fetch.bundleURI")
  utils.assertStatus(0, result, "'fetch.bundleURI' is not set after clone")
  const actualURI = result.stdout.toString().trim()
  assert.strictEqual(actualURI, this.bundleServer.bundleUri())

  result = clonedRepo.runGit("for-each-ref", "--format=%(refname)", "refs/bundles/*")
  utils.assertStatus(0, result, "git for-each-ref failed")

  const bundleRefs = result.stdout.toString().split("\n").filter(function(line) {
    return line.trim() != ""
  })
  assert.strict(bundleRefs.length > 0, "No bundle refs found in the repo")
})
