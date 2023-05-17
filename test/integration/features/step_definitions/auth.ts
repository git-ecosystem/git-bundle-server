import { IntegrationBundleServerWorld } from '../support/world'
import { Given } from '@cucumber/cucumber'
import * as fs from 'fs'
import * as path from 'path'
import * as crypto from 'crypto';

Given('an auth config with username {string} and password {string}',
  async function (this: IntegrationBundleServerWorld, user: string, pass: string) {
    const config = {
      "mode": "fixed",
      "parameters": {
        "username": user,
        "passwordHash": crypto.createHash('sha256').update(pass).digest('hex')
      }
    }
    fs.writeFileSync(path.join(this.trashDirectory, "auth-config.json"), JSON.stringify(config))
  }
)

Given('the bundle web server was started at port {int} with auth config',
  async function (this: IntegrationBundleServerWorld, port: number) {
    await this.bundleServer.startWebServer(port, path.join(this.trashDirectory, "auth-config.json"))
  }
)
