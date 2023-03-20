const common = {
  requireModule: ['ts-node/register'],
  require: ['features/**/*.ts'],
  publishQuiet: true,
  format: ['progress'],
  formatOptions: {
    snippetInterface: 'async-await'
  },
  worldParameters: {
    bundleServerCommand: '../../bin/git-bundle-server',
    bundleWebServerCommand: '../../bin/git-bundle-web-server',
    trashDirectoryBase: '../../_test/e2e'
  }
}

module.exports = {
  default: {
    ...common,
    tags: 'not @slow',
  },
  offline: {
    ...common,
    tags: 'not @online',
  },
  all: {
    ...common
  }
}
