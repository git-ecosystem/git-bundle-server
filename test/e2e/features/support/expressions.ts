import { defineParameterType } from '@cucumber/cucumber'

defineParameterType({
  name: 'boolean',
  regexp: /are|are not/,
  transformer: s => s == "are"
})
