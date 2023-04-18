import { When } from "@cucumber/cucumber"
import { IntegrationBundleServerWorld } from '../support/world'

When('I run the bundle server CLI command {string}', async function (this: IntegrationBundleServerWorld, command: string) {
  this.runCommand(command)
})
