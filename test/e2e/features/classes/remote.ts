export class RemoteRepo {
  remoteUri: string
  root: string

  constructor(url: string) {
    this.remoteUri = url
    this.root = ""
  }
