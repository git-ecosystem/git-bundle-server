import * as path from 'path'
import * as utils from '../support/utils'
import * as child_process from 'child_process'

export class RemoteRepo {
  isLocal: boolean
  remoteUri: string
  root: string

  constructor(isLocal: boolean, urlOrPath: string, mainBranch?: string) {
    this.isLocal = isLocal
    if (!this.isLocal) {
      // Not a bare repo on the filesystem
      this.remoteUri = urlOrPath
      this.root = ""
    } else {
      // Bare repo on the filesystem - need to initialize
      if (urlOrPath.startsWith("file://")) {
        this.remoteUri = urlOrPath
        this.root = urlOrPath.substring(7)
      } else if (path.isAbsolute(urlOrPath)) {
        this.remoteUri = `file://${urlOrPath}`
        this.root = urlOrPath
      } else {
        throw new Error("'urlOrPath' must be a 'file://' URL or absolute path")
      }

      utils.assertStatus(0, utils.runGit("init", "--bare", this.root))
      if (mainBranch) {
        utils.assertStatus(0, utils.runGit("-C", this.root, "symbolic-ref", "HEAD", `refs/heads/${mainBranch}`))
      }
    }
  }

  getBranchTipOid(branch: string): string {
    if (!this.isLocal) {
      throw new Error("Logged branch tips are only available for local custom remotes")
    }
    const result = child_process.spawnSync("cat", [`refs/heads/${branch}`], { shell: true, cwd: this.root })
    utils.assertStatus(0, result)
    return result.stdout.toString().trim()
  }
}
