import * as assert from 'assert'
import * as utils from '../support/utils'
import { randomBytes } from 'crypto'
import { BundleServerWorld, User } from '../support/world'
import { Given, When, Then } from '@cucumber/cucumber'

/**
 * Steps relating to the repository clones that users work with directly. Since
 * these steps represent the actions a user takes, the majority of end-to-end
 * test steps will live here.
 */

Given('another user pushed {int} commits to {string}', async function (this: BundleServerWorld, commitNum: number, branch: string) {
  const clonedRepo = this.getRepoAtBranch(User.Another, branch)

  for (let i = 0; i < commitNum; i++) {
    utils.assertStatus(0, clonedRepo.runShell(`echo ${randomBytes(16).toString('hex')} >README.md`))
    utils.assertStatus(0, clonedRepo.runGit("add", "README.md"))
    utils.assertStatus(0, clonedRepo.runGit("commit", "-m", `test ${i + 1}`))
  }
  utils.assertStatus(0, clonedRepo.runGit("push", "origin", branch))
})

Given('another user removed {int} commits and added {int} commits to {string}',
  async function (this: BundleServerWorld, removeCommits: number, addCommits: number, branch: string) {
    const clonedRepo = this.getRepoAtBranch(User.Another, branch)

    // First, reset
    utils.assertStatus(0, clonedRepo.runGit("reset", "--hard", `HEAD~${removeCommits}`))

    // Then, add new commits
    for (let i = 0; i < addCommits; i++) {
      utils.assertStatus(0, clonedRepo.runShell(`echo ${randomBytes(16).toString('hex')} >README.md`))
      utils.assertStatus(0, clonedRepo.runGit("add", "README.md"))
      utils.assertStatus(0, clonedRepo.runGit("commit", "-m", `test ${i + 1}`))
    }

    // Finally, force push
    utils.assertStatus(0, clonedRepo.runGit("push", "-f", "origin", branch))
  }
)

Given('I cloned from the remote repo with a bundle URI', async function (this: BundleServerWorld) {
  const user = User.Me
  this.cloneRepositoryFor(user, this.bundleServer.bundleUri())
  utils.assertStatus(0, this.getRepo(user).cloneResult)
})

When('I clone from the remote repo with a bundle URI', async function (this: BundleServerWorld) {
  this.cloneRepositoryFor(User.Me, this.bundleServer.bundleUri())
})

When('another developer clones from the remote repo without a bundle URI', async function (this: BundleServerWorld) {
  this.cloneRepositoryFor(User.Another)
})

When('I fetch from the remote', async function (this: BundleServerWorld) {
  const clonedRepo = this.getRepo(User.Me)
  utils.assertStatus(0, clonedRepo.runGit("fetch", "origin"))
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

Then('I am up-to-date with {string}', async function (this: BundleServerWorld, branch: string) {
  const clonedRepo = this.getRepo(User.Me)
  const result = clonedRepo.runGit("rev-parse", `refs/remotes/origin/${branch}`)
  utils.assertStatus(0, result)
  const actualOid = result.stdout.toString().trim()
  const expectedOid = this.remote?.getBranchTipOid(branch)
  assert.strictEqual(actualOid, expectedOid, `branch '${branch}' is not up-to-date`)
})

Then('my repo\'s bundles {boolean} up-to-date with {string}',
  async function (this: BundleServerWorld, expectedUpToDate: boolean, branch: string) {
    const clonedRepo = this.getRepo(User.Me)
    const result = clonedRepo.runGit("rev-parse", `refs/bundles/${branch}`)
    utils.assertStatus(0, result)
    const actualOid = result.stdout.toString().trim()
    const expectedOid = this.remote?.getBranchTipOid(branch)

    if (expectedUpToDate) {
      assert.strictEqual(actualOid, expectedOid, `bundle ref for '${branch}' is not up-to-date`)
    } else {
      assert.notStrictEqual(actualOid, expectedOid, `bundle ref for '${branch}' is up-to-date, but should not be`)
    }
  }
)

Then('I compare the clone execution times', async function (this: BundleServerWorld) {
  const myClone = this.getRepo(User.Me)
  const otherClone = this.getRepo(User.Another)

  // Verify the clones succeeded
  utils.assertStatus(0, myClone.cloneResult)
  utils.assertStatus(0, otherClone.cloneResult)

  console.log(`\nClone execution time for ${this.remote!.remoteUri}: ${(myClone.cloneTimeMs / 1000).toFixed(2)}s (bundle URI) vs. ${(otherClone.cloneTimeMs / 1000).toFixed(2)}s (no bundle URI)`)
})
