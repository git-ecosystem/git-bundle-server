import * as path from 'path'

export function absPath(pathParam: string): string {
  // Convert a given path (either relative to 'test/e2e/' or absolute) to an
  // absolute path
  if (!path.isAbsolute(pathParam)) {
    return path.resolve(__dirname, "../..", pathParam)
  } else {
    return pathParam
  }
}
