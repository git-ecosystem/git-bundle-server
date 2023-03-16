import { setWorldConstructor, World, IWorldOptions } from '@cucumber/cucumber'
import { randomUUID } from 'crypto'
import { RemoteRepo } from '../classes/remote'
import * as utils from './utils'
import * as fs from 'fs'
import * as path from 'path'
import { ClonedRepository } from '../classes/repository'
import { BundleServer } from '../classes/bundleServer'

export enum User {
  Me = 1,
  Another,
}

interface BundleServerParameters {
  bundleServerCommand: string
  bundleWebServerCommand: string
  trashDirectoryBase: string
}

export class BundleServerWorld extends World<BundleServerParameters> {
  // Internal variables
  trashDirectory: string

  // Bundle server
  bundleServer: BundleServer
  remote: RemoteRepo | undefined

  // Users
  repoMap: Map<User, ClonedRepository>

  constructor(options: IWorldOptions<BundleServerParameters>) {
    super(options)

    this.repoMap = new Map<User, ClonedRepository>()

    // Set up the trash directory
    this.trashDirectory = path.join(utils.absPath(this.parameters.trashDirectoryBase), randomUUID())
    fs.mkdirSync(this.trashDirectory, { recursive: true });

    // Set up the bundle server
    this.bundleServer = new BundleServer(utils.absPath(this.parameters.bundleServerCommand),
      utils.absPath(this.parameters.bundleWebServerCommand))
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

  cleanup(): void {
    this.bundleServer.cleanup()

    // Delete the trash directory
    fs.rmSync(this.trashDirectory, { recursive: true })
  }
}

setWorldConstructor(BundleServerWorld)
