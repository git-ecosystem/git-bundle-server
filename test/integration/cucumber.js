const common = {
  requireModule: ['ts-node/register'],
  require: [`${__dirname}/features/**/*.ts`, `${__dirname}/../shared/features/**/*.ts`],
  publishQuiet: true,
  format: ['progress'],
  formatOptions: {
    snippetInterface: 'async-await'
  },
  worldParameters: {
    bundleServerCommand: `${__dirname}/../../bin/git-bundle-server`,
    bundleWebServerCommand: `${__dirname}/../../bin/git-bundle-web-server`,
    trashDirectoryBase: `${__dirname}/../../_test/integration`
  }
}

module.exports = {
  default: {
    ...common,
  },
  offline: {
    ...common,
    tags: 'not @online',
  },
  ci: {
    ...common,
    tags: 'not @daemon',
  },
}
