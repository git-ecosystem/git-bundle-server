import * as path from 'path'
import * as assert from 'assert'
import * as child_process from 'child_process'

export function runGit(...args: string[]): child_process.SpawnSyncReturns<Buffer> {
  return child_process.spawnSync("git", args)
}

export function absPath(pathParam: string): string {
  // Convert a given path (either relative to 'test/e2e/' or absolute) to an
  // absolute path
  if (!path.isAbsolute(pathParam)) {
    return path.resolve(__dirname, "../..", pathParam)
  } else {
    return pathParam
  }
}

export function assertStatus(expectedStatusCode: number, result: child_process.SpawnSyncReturns<Buffer>, message?: string): void {
  if (result.error) {
    throw result.error
  }
  assert.strictEqual(result.status, expectedStatusCode,
    `${message ?? "Invalid status code"}:\n\tstdout: ${result.stdout.toString()}\n\tstderr: ${result.stderr.toString()}`)
}
