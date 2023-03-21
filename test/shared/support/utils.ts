import * as assert from 'assert'
import * as child_process from 'child_process'
import * as path from 'path'

const bundleRoot = `${process.env.HOME}/git-bundle-server`

export function absPath(pathParam: string): string {
  // Convert a given path (either relative to the top-level project directory or
  // absolute) to an absolute path
  if (!path.isAbsolute(pathParam)) {
    return path.resolve(__dirname, "../../../", pathParam)
  } else {
    return pathParam
  }
}

export function runGit(...args: string[]): child_process.SpawnSyncReturns<Buffer> {
    return child_process.spawnSync("git", args)
}

export function assertStatus(expectedStatusCode: number, result: child_process.SpawnSyncReturns<Buffer>, message?: string): void {
    if (result.error) {
      console.log('error: ', result.error)
      throw result.error
    }
    assert.strictEqual(result.status, expectedStatusCode,
      `${message ?? "Invalid status code"}:\n\tstdout: ${result.stdout.toString()}\n\tstderr: ${result.stderr.toString()}`)
}

export function wwwPath(): string {
  return path.resolve(bundleRoot, "www")
}

export function repoRoot(pathParam: string): string {
  if (!path.isAbsolute(pathParam)) {
    return path.resolve(bundleRoot, "git", pathParam)
  } else {
    return pathParam
  }
}

export function routesPath(): string {
  return path.resolve(bundleRoot, "routes")
}
