import * as child_process from 'child_process'

export enum DaemonState {
  Running = 0,
  NotRunning = 1
}

export function getLaunchDDaemonState(): DaemonState {
  var regex = new RegExp(/^\s*state = (.*)$/, "m")

  var user = child_process.spawnSync('id', ['-u']).stdout.toString().trim()
  var cmdResult = child_process.spawnSync('launchctl',
    ['print', `user/${user}/com.github.gitbundleserver`])

  var state = parseOutput(cmdResult.stdout.toString(), regex)

  switch(state)
  {
    case 'running':
      return DaemonState.Running
    case 'not running':
    case 'not started':
      return DaemonState.NotRunning
    default:
      throw new Error(`launchd daemon state ${state} not recognized`)
  }
}

export function getSystemDDaemonState(): DaemonState {
  var regex = new RegExp(/^\s*Active: (.*)(?= \()/, "m")

  var user = child_process.spawnSync('id', ['-u']).stdout.toString()
  var cmdResult = child_process.spawnSync('systemctl',
      ['status', '--user', user, 'com.github.gitbundleserver'])

  var state = parseOutput(cmdResult.stdout.toString(), regex)

  switch(state)
  {
    case 'active':
      return DaemonState.Running
    case 'inactive':
    case 'not started':
      return DaemonState.NotRunning
    default:
      throw new Error(`systemd daemon state ${state} not recognized`)
  }
}

function parseOutput(stdout: string, regex: RegExp): string {
  var potentialMatchParts = stdout.match(regex)
  if (potentialMatchParts) {
    return potentialMatchParts[1]
  }
  else {
    return 'not started'
  }
}
