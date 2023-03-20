import { BundleServerWorld } from '../support/world'
import { After } from '@cucumber/cucumber'

/**
 * Steps handling operations that are common across tests.
 */

After(function (this: BundleServerWorld) {
  this.cleanup()
});
