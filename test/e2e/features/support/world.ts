import { setWorldConstructor, IWorldOptions } from '@cucumber/cucumber'
import * as utils from '../../../shared/support/utils'
import { ClonedRepository } from '../../../shared/classes/repository'
import { BundleServerWorldBase, BundleServerParameters } from '../../../shared/support/world'

export enum User {
  Me = 1,
  Another,
}

export class EndToEndBundleServerWorld extends BundleServerWorldBase {
  // Users
  repoMap: Map<User, ClonedRepository>

  constructor(options: IWorldOptions<BundleServerParameters>) {
    super(options)

    this.repoMap = new Map<User, ClonedRepository>()
  }

  cloneRepositoryFor(user: User, bundleUri?: string): void {
    if (!this.remote) {
      throw new Error("Remote repository is not initialized")
    }

    const repoRoot = `${this.trashDirectory}/${User[user]}`
    this.repoMap.set(user, new ClonedRepository(this.remote, repoRoot, bundleUri))
  }

  getRepo(user: User): ClonedRepository {
    const repo = this.repoMap.get(user)
    if (!repo) {
      throw new Error("Cloned repository has not been initialized")
    }

    return repo
  }

  getRepoAtBranch(user: User, branch: string): ClonedRepository {
    if (this.remote && !this.remote.isLocal) {
      throw new Error("Remote is not initialized or does not allow pushes")
    }

    if (!this.repoMap.has(user)) {
      this.cloneRepositoryFor(user)
      utils.assertStatus(0, this.getRepo(user).cloneResult)
    }

    const clonedRepo = this.getRepo(user)

    const result = clonedRepo.runGit("rev-list", "--all", "-n", "1")
    if (result.stdout.toString().trim() == "") {
      // Repo is empty, so make sure we're on the right branch
      utils.assertStatus(0, clonedRepo.runGit("branch", "-m", branch))
    } else {
      utils.assertStatus(0, clonedRepo.runGit("switch", branch))
      utils.assertStatus(0, clonedRepo.runGit("pull", "origin", branch))
    }

    return clonedRepo
  }
}

setWorldConstructor(EndToEndBundleServerWorld)
