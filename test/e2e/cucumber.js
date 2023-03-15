module.exports = {
  default: {
    requireModule: ['ts-node/register'],
    require: ['features/**/*.ts'],
    publishQuiet: true,
    format: ['progress'],
    formatOptions: {
      snippetInterface: 'async-await'
    },
  }
}
